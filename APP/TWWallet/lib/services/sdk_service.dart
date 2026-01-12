import 'dart:convert';
import 'package:did_sdk_module/did_key_gen.dart';
import 'package:did_sdk_module/openid_vc_vp.dart';

/// SDK Response wrapper
class SDKResponse {
  final String code;
  final String message;
  final dynamic data;

  SDKResponse({required this.code, required this.message, this.data});

  bool get isSuccess => code == '0';

  factory SDKResponse.fromJson(String jsonString) {
    final json = jsonDecode(jsonString) as Map<String, dynamic>;
    return SDKResponse(
      code: json['code']?.toString() ?? '1',
      message: json['message']?.toString() ?? '',
      data: json['data'],
    );
  }
}

/// SDK Exception
class SDKException implements Exception {
  final String code;
  final String message;

  SDKException(this.code, this.message);

  @override
  String toString() => 'SDKException [$code]: $message';
}

/// Service for interacting with the DID SDK
class SDKService {
  static const sdkVersion = '0.0.1';

  final DIDKeyGenerator _didKeyGenerator = DIDKeyGenerator();
  final OpenidVcVp _openidVcVp = OpenidVcVp();

  // ============ Version ============

  String getVersion() => sdkVersion;

  // ============ Key Management ============

  /// Initialize KeyXentic module
  Future<void> initKx({
    required String keyTag,
    required String type,
    required String pin,
  }) async {
    final response = await _didKeyGenerator.initKx(keyTag, type, pin);
    _validateResponse(response);
  }

  /// Load existing key pair from storage (for app restart)
  /// Returns true if key was loaded successfully
  Future<bool> loadExistingKey() async {
    return await DIDKeyGenerator.loadExistingKey();
  }

  /// Generate P-256 public key in JWK format
  Future<Map<String, dynamic>> generateKey() async {
    final response = await _didKeyGenerator.generateKeyKx();
    final parsed = SDKResponse.fromJson(response);
    if (!parsed.isSuccess) {
      throw SDKException(parsed.code, parsed.message);
    }
    return parsed.data['publicKey'] as Map<String, dynamic>;
  }

  /// Generate DID document from public key JWK
  Future<Map<String, dynamic>> generateDID(Map<String, dynamic> publicKey) async {
    final response = await _didKeyGenerator.generateDID(jsonEncode(publicKey));
    final parsed = SDKResponse.fromJson(response);
    if (!parsed.isSuccess) {
      throw SDKException(parsed.code, parsed.message);
    }
    return parsed.data as Map<String, dynamic>;
  }

  // ============ Credential Operations (OID4VCI) ============

  /// Apply for a Verifiable Credential from an issuer
  Future<Map<String, dynamic>> applyVC({
    required Map<String, dynamic> didDocument,
    required String qrCode,
    required String otp,
  }) async {
    final response = await _openidVcVp.applyVCKx(
      jsonEncode(didDocument),
      qrCode,
      otp,
    );
    final parsed = SDKResponse.fromJson(response);
    if (!parsed.isSuccess) {
      throw SDKException(parsed.code, parsed.message);
    }
    return parsed.data as Map<String, dynamic>;
  }

  /// Decode a Verifiable Credential JWT
  Map<String, dynamic> decodeVC(String credential) {
    final response = _openidVcVp.decodeVC(credential);
    final parsed = SDKResponse.fromJson(response);
    if (!parsed.isSuccess) {
      throw SDKException(parsed.code, parsed.message);
    }
    return parsed.data as Map<String, dynamic>;
  }

  /// Verify a Verifiable Credential online
  Future<Map<String, dynamic>> verifyVC({
    required String credential,
    required Map<String, dynamic> didDocument,
    required String frontUrl,
  }) async {
    final response = await _openidVcVp.verifyVC(
      credential,
      jsonEncode(didDocument),
      frontUrl,
    );
    final parsed = SDKResponse.fromJson(response);
    if (!parsed.isSuccess) {
      throw SDKException(parsed.code, parsed.message);
    }
    return parsed.data as Map<String, dynamic>;
  }

  /// Verify a Verifiable Credential offline
  Future<Map<String, dynamic>> verifyVCOffline({
    required String credential,
    required Map<String, dynamic> didDocument,
    required List<dynamic> issuerList,
    required List<dynamic> vcStatusList,
  }) async {
    final response = await _openidVcVp.verifyVCOffline(
      credential,
      jsonEncode(didDocument),
      issuerList,
      vcStatusList,
    );
    final parsed = SDKResponse.fromJson(response);
    if (!parsed.isSuccess) {
      throw SDKException(parsed.code, parsed.message);
    }
    return parsed.data as Map<String, dynamic>;
  }

  // ============ Presentation Operations (OID4VP) ============

  /// Parse a VP request QR code
  Future<Map<String, dynamic>> parseVPQrcode({
    required String qrCode,
    required String frontUrl,
  }) async {
    final response = await _openidVcVp.parseVPQrcode(qrCode, frontUrl);
    final parsed = SDKResponse.fromJson(response);
    if (!parsed.isSuccess) {
      throw SDKException(parsed.code, parsed.message);
    }
    return parsed.data as Map<String, dynamic>;
  }

  /// Generate a Verifiable Presentation
  Future<void> generateVP({
    required Map<String, dynamic> didDocument,
    required String requestToken,
    required List<Map<String, dynamic>> vcs,
    String customData = '',
  }) async {
    final response = await _openidVcVp.generateVPKx(
      jsonEncode(didDocument),
      requestToken,
      vcs,
      customData,
    );
    _validateResponse(response);
  }

  /// Generate VP for NFC transmission
  Future<String> generateVPNFC({
    required Map<String, dynamic> didDocument,
    required String vc,
  }) async {
    final response = await _openidVcVp.generateVPNFC(
      jsonEncode(didDocument),
      vc,
    );
    final parsed = SDKResponse.fromJson(response);
    if (!parsed.isSuccess) {
      throw SDKException(parsed.code, parsed.message);
    }
    return parsed.data.toString();
  }

  /// Verify VP received via NFC
  Future<Map<String, dynamic>> verifyVPNFC(String vp) async {
    final response = await _openidVcVp.verifyVPNFC(vp);
    final parsed = SDKResponse.fromJson(response);
    if (!parsed.isSuccess) {
      throw SDKException(parsed.code, parsed.message);
    }
    return parsed.data as Map<String, dynamic>;
  }

  // ============ Offline Data ============

  /// Download issuer trust list for offline verification
  Future<List<dynamic>> downloadIssuerList(String url) async {
    final response = await _openidVcVp.downloadIssList(url);
    final parsed = SDKResponse.fromJson(response);
    if (!parsed.isSuccess) {
      throw SDKException(parsed.code, parsed.message);
    }
    return parsed.data as List<dynamic>;
  }

  /// Download VC status lists for offline verification
  Future<List<dynamic>> downloadAllVCList({
    required List<String> vcs,
    required List<dynamic> existingVcList,
  }) async {
    final response = await _openidVcVp.downloadAllVCList(vcs, existingVcList);
    final parsed = SDKResponse.fromJson(response);
    if (!parsed.isSuccess) {
      throw SDKException(parsed.code, parsed.message);
    }
    return parsed.data as List<dynamic>;
  }

  // ============ Transfer ============

  /// Transfer a credential to another device
  Future<Map<String, dynamic>> transferVC({
    required Map<String, dynamic> didDocument,
    required String credential,
  }) async {
    final response = await _openidVcVp.transferVC(
      jsonEncode(didDocument),
      credential,
    );
    final parsed = SDKResponse.fromJson(response);
    if (!parsed.isSuccess) {
      throw SDKException(parsed.code, parsed.message);
    }
    return parsed.data as Map<String, dynamic>;
  }

  // ============ Generic Requests ============

  /// Send a generic HTTP request
  Future<Map<String, dynamic>> sendRequest({
    required String url,
    required String type,
    required String body,
  }) async {
    final response = await _openidVcVp.sendRequest(url, type, body);
    final parsed = SDKResponse.fromJson(response);
    if (!parsed.isSuccess) {
      throw SDKException(parsed.code, parsed.message);
    }
    return parsed.data as Map<String, dynamic>;
  }

  /// Send a JWT-authenticated request
  Future<Map<String, dynamic>> sendJWTRequest({
    required String url,
    required String payload,
    required Map<String, dynamic> didDocument,
  }) async {
    final response = await _openidVcVp.sendJWTRequest(
      url,
      payload,
      jsonEncode(didDocument),
    );
    final parsed = SDKResponse.fromJson(response);
    if (!parsed.isSuccess) {
      throw SDKException(parsed.code, parsed.message);
    }
    return parsed.data as Map<String, dynamic>;
  }

  // ============ Helpers ============

  void _validateResponse(String response) {
    final parsed = SDKResponse.fromJson(response);
    if (!parsed.isSuccess) {
      throw SDKException(parsed.code, parsed.message);
    }
  }
}
