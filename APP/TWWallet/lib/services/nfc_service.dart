import 'dart:convert';
import 'package:nfc_manager/nfc_manager.dart';

/// NFC operation result
class NFCResult {
  final bool success;
  final String? data;
  final String? error;

  NFCResult({required this.success, this.data, this.error});
}

/// Service for NFC operations
class NFCService {
  /// Check if NFC is available on this device
  Future<bool> isNFCAvailable() async {
    return await NfcManager.instance.isAvailable();
  }

  /// Write VP data to NFC tag (for presenting credentials)
  Future<NFCResult> writeVPToNFC(String vpData) async {
    try {
      final isAvailable = await isNFCAvailable();
      if (!isAvailable) {
        return NFCResult(
          success: false,
          error: '此裝置不支援 NFC 功能',
        );
      }

      // Start NFC session
      await NfcManager.instance.startSession(
        onDiscovered: (NfcTag tag) async {
          try {
            // Try to get NDEF instance
            final ndef = Ndef.from(tag);
            if (ndef == null) {
              throw Exception('此 NFC 標籤不支援 NDEF 格式');
            }

            if (!ndef.isWritable) {
              throw Exception('此 NFC 標籤不可寫入');
            }

            // Create NDEF message with VP data
            final ndefMessage = NdefMessage([
              NdefRecord.createText(vpData),
            ]);

            // Check if the message fits
            if (ndef.maxSize < ndefMessage.byteLength) {
              throw Exception('資料太大，無法寫入此 NFC 標籤');
            }

            // Write the message
            await ndef.write(ndefMessage);

            // Stop the session
            await NfcManager.instance.stopSession();
          } catch (e) {
            await NfcManager.instance.stopSession(errorMessage: e.toString());
            rethrow;
          }
        },
      );

      return NFCResult(success: true, data: 'VP 已成功寫入 NFC');
    } catch (e) {
      return NFCResult(success: false, error: e.toString());
    }
  }

  /// Read VP data from NFC tag (for verifying credentials)
  Future<NFCResult> readVPFromNFC() async {
    try {
      final isAvailable = await isNFCAvailable();
      if (!isAvailable) {
        return NFCResult(
          success: false,
          error: '此裝置不支援 NFC 功能',
        );
      }

      String? vpData;

      await NfcManager.instance.startSession(
        onDiscovered: (NfcTag tag) async {
          try {
            // Try to get NDEF instance
            final ndef = Ndef.from(tag);
            if (ndef == null) {
              throw Exception('此 NFC 標籤不支援 NDEF 格式');
            }

            // Read the cached message or read from tag
            final ndefMessage = await ndef.read();

            if (ndefMessage.records.isEmpty) {
              throw Exception('NFC 標籤中沒有資料');
            }

            // Extract text from the first record
            final record = ndefMessage.records.first;
            vpData = _decodeNdefTextRecord(record);

            // Stop the session
            await NfcManager.instance.stopSession();
          } catch (e) {
            await NfcManager.instance.stopSession(errorMessage: e.toString());
            rethrow;
          }
        },
      );

      if (vpData != null) {
        return NFCResult(success: true, data: vpData);
      } else {
        return NFCResult(success: false, error: '無法讀取 NFC 資料');
      }
    } catch (e) {
      return NFCResult(success: false, error: e.toString());
    }
  }

  /// Start NFC session for writing VP
  Future<void> startWriteSession({
    required String vpData,
    required Function(bool success, String message) onResult,
  }) async {
    try {
      final isAvailable = await isNFCAvailable();
      if (!isAvailable) {
        onResult(false, '此裝置不支援 NFC 功能');
        return;
      }

      await NfcManager.instance.startSession(
        alertMessage: '請將裝置靠近 NFC 讀卡機',
        onDiscovered: (NfcTag tag) async {
          try {
            final ndef = Ndef.from(tag);
            if (ndef == null) {
              onResult(false, '此 NFC 標籤不支援 NDEF 格式');
              await NfcManager.instance.stopSession(errorMessage: '不支援的標籤格式');
              return;
            }

            if (!ndef.isWritable) {
              onResult(false, '此 NFC 標籤不可寫入');
              await NfcManager.instance.stopSession(errorMessage: '標籤不可寫入');
              return;
            }

            final ndefMessage = NdefMessage([
              NdefRecord.createText(vpData),
            ]);

            if (ndef.maxSize < ndefMessage.byteLength) {
              onResult(false, '資料太大，無法寫入');
              await NfcManager.instance.stopSession(errorMessage: '資料太大');
              return;
            }

            await ndef.write(ndefMessage);
            onResult(true, '憑證已成功傳輸');
            await NfcManager.instance.stopSession(alertMessage: '傳輸成功！');
          } catch (e) {
            onResult(false, '傳輸失敗: $e');
            await NfcManager.instance.stopSession(errorMessage: '傳輸失敗');
          }
        },
      );
    } catch (e) {
      onResult(false, '無法啟動 NFC: $e');
    }
  }

  /// Start NFC session for reading VP
  Future<void> startReadSession({
    required Function(bool success, String data) onResult,
  }) async {
    try {
      final isAvailable = await isNFCAvailable();
      if (!isAvailable) {
        onResult(false, '此裝置不支援 NFC 功能');
        return;
      }

      await NfcManager.instance.startSession(
        alertMessage: '請將裝置靠近 NFC 標籤',
        onDiscovered: (NfcTag tag) async {
          try {
            final ndef = Ndef.from(tag);
            if (ndef == null) {
              onResult(false, '此 NFC 標籤不支援 NDEF 格式');
              await NfcManager.instance.stopSession(errorMessage: '不支援的格式');
              return;
            }

            final ndefMessage = await ndef.read();

            if (ndefMessage.records.isEmpty) {
              onResult(false, 'NFC 標籤中沒有資料');
              await NfcManager.instance.stopSession(errorMessage: '沒有資料');
              return;
            }

            final record = ndefMessage.records.first;
            final vpData = _decodeNdefTextRecord(record);

            if (vpData != null) {
              onResult(true, vpData);
              await NfcManager.instance.stopSession(alertMessage: '讀取成功！');
            } else {
              onResult(false, '無法解析資料');
              await NfcManager.instance.stopSession(errorMessage: '解析失敗');
            }
          } catch (e) {
            onResult(false, '讀取失敗: $e');
            await NfcManager.instance.stopSession(errorMessage: '讀取失敗');
          }
        },
      );
    } catch (e) {
      onResult(false, '無法啟動 NFC: $e');
    }
  }

  /// Stop any ongoing NFC session
  Future<void> stopSession() async {
    try {
      await NfcManager.instance.stopSession();
    } catch (_) {
      // Ignore errors when stopping
    }
  }

  /// Decode NDEF text record to string
  String? _decodeNdefTextRecord(NdefRecord record) {
    if (record.typeNameFormat != NdefTypeNameFormat.nfcWellknown) {
      return null;
    }

    final payload = record.payload;
    if (payload.isEmpty) return null;

    // Text record format: [status byte] [language code] [text]
    final statusByte = payload[0];
    final languageCodeLength = statusByte & 0x3F;
    final isUtf16 = (statusByte & 0x80) != 0;

    if (payload.length < 1 + languageCodeLength) return null;

    final textBytes = payload.sublist(1 + languageCodeLength);

    if (isUtf16) {
      // UTF-16 encoding
      return String.fromCharCodes(textBytes);
    } else {
      // UTF-8 encoding
      return utf8.decode(textBytes);
    }
  }
}
