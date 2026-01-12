import 'dart:convert';
import 'dart:math';
import 'dart:typed_data';
import 'package:pointycastle/export.dart';
import 'package:convert/convert.dart';
import 'package:base_x/base_x.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'utils/utils.dart';

/// Pure Dart implementation for DID key generation
/// Uses pointycastle for P-256 key generation (development/testing purposes)
class DIDKeyGenerator {
  // Singleton instance for session management
  static final DIDKeyGenerator _instance = DIDKeyGenerator._internal();
  factory DIDKeyGenerator() => _instance;
  DIDKeyGenerator._internal();

  Utils utils = Utils();

  // Secure storage for persisting private key
  static const _secureStorage = FlutterSecureStorage(
    aOptions: AndroidOptions(encryptedSharedPreferences: true),
    iOptions: IOSOptions(accessibility: KeychainAccessibility.first_unlock),
  );

  static const String _privateKeyStorageKey = 'did_private_key_d';
  static const String _publicKeyXStorageKey = 'did_public_key_x';
  static const String _publicKeyYStorageKey = 'did_public_key_y';
  static const String _pinHashStorageKey = 'did_pin_hash';
  static const String _pinSaltStorageKey = 'did_pin_salt';

  // PBKDF2 parameters
  static const int _pbkdf2Iterations = 100000;
  static const int _saltLength = 32;
  static const int _keyLength = 32;

  // Store generated key pair for signing operations (in memory for session)
  // Instance variables - not accessible without valid session
  AsymmetricKeyPair<PublicKey, PrivateKey>? _keyPair;
  String? _keyTag;
  bool _isSessionValid = false;

  /// Initialize the key module (pure Dart implementation)
  ///
  /// [keyTag] Key identifier tag
  /// [type] Key type (e.g., "P256")
  /// [pin] User PIN code
  ///
  /// Returns success/error response
  Future<String> initKx(String keyTag, String type, String pin) async {
    try {
      _keyTag = keyTag;

      // Check if PIN hash already exists (returning user)
      final existingHash = await _secureStorage.read(key: _pinHashStorageKey);

      if (existingHash != null) {
        // Verify PIN against stored hash
        final isValid = await _verifyPin(pin);
        if (!isValid) {
          return utils.response(code: '2', message: 'Invalid PIN');
        }
      } else {
        // First time setup - store PIN hash with salt
        await _storePinHash(pin);
      }

      // Mark session as valid after PIN verification
      _isSessionValid = true;

      // Try to load existing key pair from secure storage
      await _loadKeyPairFromStorage();

      return utils.response(code: '0', message: 'true');
    } catch (e) {
      return utils.response(code: '1', message: e.toString());
    }
  }

  /// Store PIN hash using PBKDF2 with random salt
  Future<void> _storePinHash(String pin) async {
    final salt = _generateSalt();
    final hash = _hashPinWithPbkdf2(pin, salt);

    await _secureStorage.write(key: _pinSaltStorageKey, value: hex.encode(salt));
    await _secureStorage.write(key: _pinHashStorageKey, value: hex.encode(hash));
  }

  /// Verify PIN against stored hash
  Future<bool> _verifyPin(String pin) async {
    final storedSaltHex = await _secureStorage.read(key: _pinSaltStorageKey);
    final storedHashHex = await _secureStorage.read(key: _pinHashStorageKey);

    if (storedSaltHex == null || storedHashHex == null) {
      return false;
    }

    final salt = Uint8List.fromList(hex.decode(storedSaltHex));
    final computedHash = _hashPinWithPbkdf2(pin, salt);
    final storedHash = Uint8List.fromList(hex.decode(storedHashHex));

    // Constant-time comparison to prevent timing attacks
    return _constantTimeCompare(computedHash, storedHash);
  }

  /// Constant-time comparison to prevent timing attacks
  bool _constantTimeCompare(Uint8List a, Uint8List b) {
    if (a.length != b.length) return false;
    int result = 0;
    for (int i = 0; i < a.length; i++) {
      result |= a[i] ^ b[i];
    }
    return result == 0;
  }

  /// Generate cryptographically secure random salt
  Uint8List _generateSalt() {
    final random = Random.secure();
    return Uint8List.fromList(
      List<int>.generate(_saltLength, (_) => random.nextInt(256)),
    );
  }

  /// Hash PIN using PBKDF2 with SHA256
  Uint8List _hashPinWithPbkdf2(String pin, Uint8List salt) {
    final pbkdf2 = PBKDF2KeyDerivator(HMac(SHA256Digest(), 64))
      ..init(Pbkdf2Parameters(salt, _pbkdf2Iterations, _keyLength));

    final pinBytes = Uint8List.fromList(utf8.encode(pin));
    return pbkdf2.process(pinBytes);
  }

  /// Load key pair from secure storage if it exists
  Future<void> _loadKeyPairFromStorage() async {
    try {
      final dHex = await _secureStorage.read(key: _privateKeyStorageKey);
      final xHex = await _secureStorage.read(key: _publicKeyXStorageKey);
      final yHex = await _secureStorage.read(key: _publicKeyYStorageKey);

      if (dHex != null && xHex != null && yHex != null) {
        final ecParams = ECCurve_secp256r1();

        // Reconstruct private key
        final d = BigInt.parse(dHex, radix: 16);
        final privateKey = ECPrivateKey(d, ecParams);

        // Reconstruct public key
        final x = BigInt.parse(xHex, radix: 16);
        final y = BigInt.parse(yHex, radix: 16);
        final q = ecParams.curve.createPoint(x, y);
        final publicKey = ECPublicKey(q, ecParams);

        _keyPair = AsymmetricKeyPair(publicKey, privateKey);
      }
    } catch (e) {
      // Key not found or corrupted, will generate new one
      _keyPair = null;
    }
  }

  /// Save key pair to secure storage
  Future<void> _saveKeyPairToStorage(AsymmetricKeyPair<PublicKey, PrivateKey> keyPair) async {
    final privateKey = keyPair.privateKey as ECPrivateKey;
    final publicKey = keyPair.publicKey as ECPublicKey;

    // Store private key d value
    final dHex = privateKey.d!.toRadixString(16);
    await _secureStorage.write(key: _privateKeyStorageKey, value: dHex);

    // Store public key x and y values
    final xHex = publicKey.Q!.x!.toBigInteger()!.toRadixString(16);
    final yHex = publicKey.Q!.y!.toBigInteger()!.toRadixString(16);
    await _secureStorage.write(key: _publicKeyXStorageKey, value: xHex);
    await _secureStorage.write(key: _publicKeyYStorageKey, value: yHex);
  }

  /// Generate P-256 key pair and return public key in JWK format (pure Dart)
  ///
  /// Returns response with `publicKey` field on success
  Future<String> generateKeyKx() async {
    try {
      // Check if key pair already exists in memory or storage
      if (_keyPair == null) {
        await _loadKeyPairFromStorage();
      }

      // If still no key pair, generate a new one
      if (_keyPair == null) {
        final keyPair = _generateP256KeyPair();
        _keyPair = keyPair;

        // Persist the key pair
        await _saveKeyPairToStorage(keyPair);
      }

      // Convert public key to JWK format
      final publicKey = _keyPair!.publicKey as ECPublicKey;
      final jwk = _publicKeyToJwk(publicKey);

      final data = {
        'publicKey': jwk,
      };
      return utils.response(code: '0', message: 'SUCCESS', data: data);
    } catch (e) {
      return utils.response(code: '1', message: e.toString());
    }
  }

  /// Sign data using the stored private key
  /// Requires valid session (PIN must be verified via initKx first)
  Future<String> signKx(Uint8List data) async {
    try {
      // Verify session is valid (PIN was verified)
      if (!_isSessionValid) {
        return utils.response(code: '2', message: 'Session not initialized. Call initKx first.');
      }

      // Ensure key pair is loaded
      if (_keyPair == null) {
        await _loadKeyPairFromStorage();
      }

      if (_keyPair == null) {
        return utils.response(code: '1', message: 'Key not initialized');
      }

      final privateKey = _keyPair!.privateKey as ECPrivateKey;
      final signature = _signWithP256(data, privateKey);

      return utils.response(code: '0', message: 'SUCCESS', data: {
        'signature': base64Url.encode(signature),
      });
    } catch (e) {
      return utils.response(code: '1', message: e.toString());
    }
  }

  /// Generate DID document from public key JWK
  ///
  /// [publicKeyJwk] is a JSON string of the JWK public key
  ///
  /// Returns W3C DID compliant document on success
  Future<String> generateDID(String publicKeyJwk) async {
    try {
      Map<String, dynamic> publicKey = jsonDecode(publicKeyJwk);
      String did = generateDidFromJwk(publicKeyJwk);

      final context = [
        "https://www.w3.org/ns/did/v1",
        "https://w3id.org/security/suites/jws-2020/v1"
      ];

      List<Map<String, dynamic>> verificationMethodEntry = [
        {
          'id': 'did:key:$did#$did',
          'type': 'JsonWebKey2020',
          'controller': 'did:key:$did',
          'publicKeyJwk': publicKey
        }
      ];

      Map<String, dynamic> response = {
        '@context': context,
        'id': 'did:key:$did',
        'verificationMethod': verificationMethodEntry
      };

      return utils.response(code: '0', message: 'SUCCESS', data: response);
    } catch (e) {
      return utils.response(code: '1', message: e.toString());
    }
  }

  /// Generate P-256 key pair using pointycastle
  AsymmetricKeyPair<PublicKey, PrivateKey> _generateP256KeyPair() {
    final ecParams = ECCurve_secp256r1();
    final keyParams = ECKeyGeneratorParameters(ecParams);

    final secureRandom = _getSecureRandom();
    final keyGenerator = ECKeyGenerator()
      ..init(ParametersWithRandom(keyParams, secureRandom));

    return keyGenerator.generateKeyPair();
  }

  /// Get secure random generator
  SecureRandom _getSecureRandom() {
    final secureRandom = FortunaRandom();
    final random = Random.secure();
    final seeds = List<int>.generate(32, (_) => random.nextInt(256));
    secureRandom.seed(KeyParameter(Uint8List.fromList(seeds)));
    return secureRandom;
  }

  /// Convert EC public key to JWK format
  Map<String, dynamic> _publicKeyToJwk(ECPublicKey publicKey) {
    final q = publicKey.Q!;
    final x = _bigIntToBytes(q.x!.toBigInteger()!, 32);
    final y = _bigIntToBytes(q.y!.toBigInteger()!, 32);

    return {
      'kty': 'EC',
      'crv': 'P-256',
      'x': _base64UrlEncode(x),
      'y': _base64UrlEncode(y),
    };
  }

  /// Sign data with P-256 private key
  Uint8List _signWithP256(Uint8List data, ECPrivateKey privateKey) {
    final signer = ECDSASigner(SHA256Digest())
      ..init(true, PrivateKeyParameter<ECPrivateKey>(privateKey));

    final signature = signer.generateSignature(data) as ECSignature;

    // Convert signature to fixed-length format (R || S, each 32 bytes)
    final r = _bigIntToBytes(signature.r, 32);
    final s = _bigIntToBytes(signature.s, 32);

    return Uint8List.fromList([...r, ...s]);
  }

  /// Convert BigInt to fixed-length bytes
  Uint8List _bigIntToBytes(BigInt bigInt, int length) {
    final bytes = _bigIntToUint8List(bigInt);
    if (bytes.length >= length) {
      return Uint8List.fromList(bytes.sublist(bytes.length - length));
    }
    final padded = Uint8List(length);
    padded.setRange(length - bytes.length, length, bytes);
    return padded;
  }

  Uint8List _bigIntToUint8List(BigInt bigInt) {
    final hexString = bigInt.toRadixString(16).padLeft(64, '0');
    return Uint8List.fromList(hex.decode(hexString));
  }

  /// Base64Url encode without padding
  String _base64UrlEncode(Uint8List bytes) {
    return base64Url.encode(bytes).replaceAll('=', '');
  }

  /// Convert JWK to ECPublicKey
  ECPublicKey convertJwkToECPublicKey(Map<String, dynamic> jwk) {
    final x = base64Url
        .decode(Utils().addBase64Padding(jwk['x']));
    final y = base64Url
        .decode(Utils().addBase64Padding(jwk['y']));

    final eccParams = ECCurve_secp256r1();
    final curve = eccParams.curve;
    final xBigInt = BigInt.parse(hex.encode(x), radix: 16);
    final yBigInt = BigInt.parse(hex.encode(y), radix: 16);
    final ecPoint = curve.createPoint(xBigInt, yBigInt);

    return ECPublicKey(ecPoint, eccParams);
  }

  /// Generate DID string from JWK
  String generateDidFromJwk(String jwk) {
    List<int> bytes = utf8.encode(jwk);
    String hexString = hex.encode(bytes);

    // Multicodec prefix
    String hexPrefix = "d1d603";
    String combinedHex = hexPrefix + hexString;

    // Base58 encode
    const String base58Alphabet =
        "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz";
    final BaseXCodec base58 = BaseXCodec(base58Alphabet);
    Uint8List combinedHexBytes = Uint8List.fromList(hex.decode(combinedHex));
    String didKey = "z${base58.encode(combinedHexBytes)}";

    return didKey;
  }

  /// Get the current key pair (for signing operations in other modules)
  /// Requires valid session (PIN must be verified via initKx first)
  /// This will load from storage if not in memory
  static Future<AsymmetricKeyPair<PublicKey, PrivateKey>?> getKeyPairAsync() async {
    final instance = DIDKeyGenerator();
    if (!instance._isSessionValid) {
      throw Exception('Session not initialized. Call initKx with valid PIN first.');
    }
    if (instance._keyPair == null) {
      await instance._loadKeyPairFromStorage();
    }
    return instance._keyPair;
  }

  /// Synchronous getter - may return null if not loaded yet
  /// Requires valid session
  static AsymmetricKeyPair<PublicKey, PrivateKey>? getKeyPair() {
    final instance = DIDKeyGenerator();
    if (!instance._isSessionValid) {
      return null;
    }
    return instance._keyPair;
  }

  /// Check if session is valid (PIN verified)
  static bool isSessionValid() => DIDKeyGenerator()._isSessionValid;

  /// Clear all stored keys and invalidate session (for wallet reset)
  Future<void> clearKeys() async {
    await _secureStorage.delete(key: _privateKeyStorageKey);
    await _secureStorage.delete(key: _publicKeyXStorageKey);
    await _secureStorage.delete(key: _publicKeyYStorageKey);
    await _secureStorage.delete(key: _pinHashStorageKey);
    await _secureStorage.delete(key: _pinSaltStorageKey);
    _keyPair = null;
    _keyTag = null;
    _isSessionValid = false;
  }

  /// Invalidate current session (for logout without clearing keys)
  void invalidateSession() {
    _keyPair = null;
    _isSessionValid = false;
  }

  /// Load existing key pair from storage (for app restart)
  /// Note: This only loads the key, session still requires PIN verification via initKx
  /// Returns true if key exists in storage
  static Future<bool> loadExistingKey() async {
    final instance = DIDKeyGenerator();
    if (instance._keyPair != null) return true;
    await instance._loadKeyPairFromStorage();
    return instance._keyPair != null;
  }
}
