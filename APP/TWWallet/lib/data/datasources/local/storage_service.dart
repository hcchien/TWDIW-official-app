import 'dart:convert';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:hive_flutter/hive_flutter.dart';
import '../../models/credential.dart';
import '../../models/activity_log.dart';
import '../../../core/constants/app_constants.dart';

/// Service for local storage operations
class StorageService {
  static const String _credentialsBoxName = 'credentials';
  static const String _activityBoxName = 'activities';

  final FlutterSecureStorage _secureStorage;
  late Box<StoredCredential> _credentialsBox;
  late Box<ActivityLog> _activityBox;

  StorageService({FlutterSecureStorage? secureStorage})
      : _secureStorage = secureStorage ??
            const FlutterSecureStorage(
              aOptions: AndroidOptions(encryptedSharedPreferences: true),
              iOptions: IOSOptions(accessibility: KeychainAccessibility.first_unlock),
            );

  /// Initialize Hive and open boxes
  Future<void> init() async {
    await Hive.initFlutter();

    // Register adapters
    if (!Hive.isAdapterRegistered(0)) {
      Hive.registerAdapter(CredentialStatusAdapter());
    }
    if (!Hive.isAdapterRegistered(1)) {
      Hive.registerAdapter(CredentialFieldAdapter());
    }
    if (!Hive.isAdapterRegistered(2)) {
      Hive.registerAdapter(CredentialDisplayAdapter());
    }
    if (!Hive.isAdapterRegistered(3)) {
      Hive.registerAdapter(StoredCredentialAdapter());
    }
    if (!Hive.isAdapterRegistered(4)) {
      Hive.registerAdapter(ActivityTypeAdapter());
    }
    if (!Hive.isAdapterRegistered(5)) {
      Hive.registerAdapter(ActivityLogAdapter());
    }

    // Open boxes
    _credentialsBox = await Hive.openBox<StoredCredential>(_credentialsBoxName);
    _activityBox = await Hive.openBox<ActivityLog>(_activityBoxName);
  }

  // ============ Secure Storage (DID, Keys) ============

  /// Save DID document to secure storage
  Future<void> saveDIDDocument(Map<String, dynamic> didDocument) async {
    await _secureStorage.write(
      key: AppConstants.keyDIDDocument,
      value: jsonEncode(didDocument),
    );
  }

  /// Get DID document from secure storage
  Future<Map<String, dynamic>?> getDIDDocument() async {
    final value = await _secureStorage.read(key: AppConstants.keyDIDDocument);
    if (value == null) return null;
    return jsonDecode(value) as Map<String, dynamic>;
  }

  /// Check if DID exists
  Future<bool> hasDID() async {
    final value = await _secureStorage.read(key: AppConstants.keyDIDDocument);
    return value != null;
  }

  /// Save public key JWK
  Future<void> savePublicKey(Map<String, dynamic> publicKey) async {
    await _secureStorage.write(
      key: AppConstants.keyPublicKey,
      value: jsonEncode(publicKey),
    );
  }

  /// Get public key JWK
  Future<Map<String, dynamic>?> getPublicKey() async {
    final value = await _secureStorage.read(key: AppConstants.keyPublicKey);
    if (value == null) return null;
    return jsonDecode(value) as Map<String, dynamic>;
  }

  /// Mark onboarding as complete
  Future<void> setOnboardingComplete(bool complete) async {
    await _secureStorage.write(
      key: AppConstants.keyOnboardingComplete,
      value: complete.toString(),
    );
  }

  /// Check if onboarding is complete
  Future<bool> isOnboardingComplete() async {
    final value =
        await _secureStorage.read(key: AppConstants.keyOnboardingComplete);
    return value == 'true';
  }

  /// Enable/disable biometric
  Future<void> setBiometricEnabled(bool enabled) async {
    await _secureStorage.write(
      key: AppConstants.keyBiometricEnabled,
      value: enabled.toString(),
    );
  }

  /// Check if biometric is enabled
  Future<bool> isBiometricEnabled() async {
    final value =
        await _secureStorage.read(key: AppConstants.keyBiometricEnabled);
    return value == 'true';
  }

  // ============ Credentials (Hive) ============

  /// Get all stored credentials
  List<StoredCredential> getAllCredentials() {
    return _credentialsBox.values.toList();
  }

  /// Get credential by ID
  StoredCredential? getCredential(String id) {
    return _credentialsBox.get(id);
  }

  /// Save a credential
  Future<void> saveCredential(StoredCredential credential) async {
    await _credentialsBox.put(credential.id, credential);
  }

  /// Delete a credential
  Future<void> deleteCredential(String id) async {
    await _credentialsBox.delete(id);
  }

  /// Update credential status
  Future<void> updateCredentialStatus(
      String id, CredentialStatus status) async {
    final credential = _credentialsBox.get(id);
    if (credential != null) {
      await _credentialsBox.put(id, credential.copyWith(status: status));
    }
  }

  /// Clear all credentials
  Future<void> clearAllCredentials() async {
    await _credentialsBox.clear();
  }

  // ============ Activity Log (Hive) ============

  /// Get all activity logs
  List<ActivityLog> getAllActivities() {
    final activities = _activityBox.values.toList();
    activities.sort((a, b) => b.timestamp.compareTo(a.timestamp));
    return activities;
  }

  /// Add activity log entry
  Future<void> addActivity(ActivityLog activity) async {
    await _activityBox.put(activity.id, activity);
  }

  /// Get activities for a specific credential
  List<ActivityLog> getActivitiesForCredential(String credentialId) {
    return _activityBox.values
        .where((a) => a.credentialId == credentialId)
        .toList()
      ..sort((a, b) => b.timestamp.compareTo(a.timestamp));
  }

  /// Clear all activities
  Future<void> clearAllActivities() async {
    await _activityBox.clear();
  }

  // ============ Offline Data ============

  /// Save issuer list for offline verification
  Future<void> saveIssuerList(List<dynamic> issuerList) async {
    await _secureStorage.write(
      key: AppConstants.keyIssuerList,
      value: jsonEncode(issuerList),
    );
  }

  /// Get issuer list
  Future<List<dynamic>> getIssuerList() async {
    final value = await _secureStorage.read(key: AppConstants.keyIssuerList);
    if (value == null) return [];
    return jsonDecode(value) as List<dynamic>;
  }

  /// Save VC status list for offline verification
  Future<void> saveVCStatusList(List<dynamic> vcStatusList) async {
    await _secureStorage.write(
      key: AppConstants.keyVCStatusList,
      value: jsonEncode(vcStatusList),
    );
  }

  /// Get VC status list
  Future<List<dynamic>> getVCStatusList() async {
    final value = await _secureStorage.read(key: AppConstants.keyVCStatusList);
    if (value == null) return [];
    return jsonDecode(value) as List<dynamic>;
  }

  // ============ Clear All ============

  /// Clear all data (for logout/reset)
  Future<void> clearAll() async {
    await _credentialsBox.clear();
    await _activityBox.clear();
    await _secureStorage.deleteAll();
  }
}
