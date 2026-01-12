import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../services/sdk_service.dart';
import '../../data/datasources/local/storage_service.dart';
import '../../data/models/credential.dart';
import '../../data/models/activity_log.dart';
import 'package:uuid/uuid.dart';

// ============ Service Providers ============

final sdkServiceProvider = Provider<SDKService>((ref) => SDKService());

final storageServiceProvider = Provider<StorageService>((ref) => StorageService());

// ============ Identity Provider ============

class IdentityState {
  final bool isInitialized;
  final bool hasIdentity;
  final Map<String, dynamic>? didDocument;
  final bool isLoading;
  final String? error;

  const IdentityState({
    this.isInitialized = false,
    this.hasIdentity = false,
    this.didDocument,
    this.isLoading = false,
    this.error,
  });

  IdentityState copyWith({
    bool? isInitialized,
    bool? hasIdentity,
    Map<String, dynamic>? didDocument,
    bool? isLoading,
    String? error,
  }) {
    return IdentityState(
      isInitialized: isInitialized ?? this.isInitialized,
      hasIdentity: hasIdentity ?? this.hasIdentity,
      didDocument: didDocument ?? this.didDocument,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }
}

class IdentityNotifier extends StateNotifier<IdentityState> {
  final SDKService _sdkService;
  final StorageService _storageService;

  IdentityNotifier(this._sdkService, this._storageService)
      : super(const IdentityState());

  Future<void> initialize() async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      await _storageService.init();
      final didDocument = await _storageService.getDIDDocument();

      // If identity exists, load the key pair from secure storage
      if (didDocument != null) {
        await _sdkService.loadExistingKey();
      }

      state = state.copyWith(
        isInitialized: true,
        hasIdentity: didDocument != null,
        didDocument: didDocument,
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(
        isInitialized: true,
        isLoading: false,
        error: e.toString(),
      );
    }
  }

  Future<void> createIdentity(String pin) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      // Initialize KeyXentic
      await _sdkService.initKx(
        keyTag: 'tw_wallet_key',
        type: 'P256',
        pin: pin,
      );

      // Generate key
      final publicKey = await _sdkService.generateKey();
      await _storageService.savePublicKey(publicKey);

      // Generate DID
      final didDocument = await _sdkService.generateDID(publicKey);
      await _storageService.saveDIDDocument(didDocument);
      await _storageService.setOnboardingComplete(true);

      state = state.copyWith(
        hasIdentity: true,
        didDocument: didDocument,
        isLoading: false,
      );
    } on SDKException catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.message,
      );
      rethrow;
    } catch (e) {
      state = state.copyWith(
        isLoading: false,
        error: e.toString(),
      );
      rethrow;
    }
  }
}

final identityProvider =
    StateNotifierProvider<IdentityNotifier, IdentityState>((ref) {
  return IdentityNotifier(
    ref.watch(sdkServiceProvider),
    ref.watch(storageServiceProvider),
  );
});

// ============ Credentials Provider ============

class CredentialsNotifier extends StateNotifier<AsyncValue<List<StoredCredential>>> {
  final SDKService _sdkService;
  final StorageService _storageService;
  final Ref _ref;

  CredentialsNotifier(this._sdkService, this._storageService, this._ref)
      : super(const AsyncValue.loading());

  Future<void> loadCredentials() async {
    state = const AsyncValue.loading();
    try {
      final credentials = _storageService.getAllCredentials();
      state = AsyncValue.data(credentials);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }

  Future<StoredCredential> receiveCredential({
    required String qrCode,
    required String otp,
  }) async {
    final identityState = _ref.read(identityProvider);
    if (identityState.didDocument == null) {
      throw Exception('No identity found');
    }

    // Apply for VC
    final result = await _sdkService.applyVC(
      didDocument: identityState.didDocument!,
      qrCode: qrCode,
      otp: otp,
    );

    // Decode the credential
    final rawToken = result['credential'] as String;
    final decoded = _sdkService.decodeVC(rawToken);

    // Extract credential metadata
    final metadata = result['credentialMetadata'] as Map<String, dynamic>?;

    // Parse fields from decoded VC
    final vcData = decoded['vc'] as Map<String, dynamic>?;
    final credentialSubject = vcData?['credentialSubject'] as Map<String, dynamic>? ?? {};

    final fields = <CredentialField>[];
    credentialSubject.forEach((key, value) {
      if (key != 'id') {
        fields.add(CredentialField(
          key: key,
          label: _getFieldLabel(key),
          value: value?.toString() ?? '',
          isDisclosable: true,
        ));
      }
    });

    // Create stored credential
    final credential = StoredCredential(
      id: decoded['jti']?.toString() ?? const Uuid().v4(),
      rawToken: rawToken,
      credentialType: _extractCredentialType(vcData),
      issuer: decoded['iss']?.toString() ?? '',
      issuerName: metadata?['display']?[0]?['name']?.toString() ?? 'Unknown Issuer',
      issuedAt: DateTime.fromMillisecondsSinceEpoch(
        (decoded['iat'] as int? ?? 0) * 1000,
      ),
      expiresAt: DateTime.fromMillisecondsSinceEpoch(
        (decoded['exp'] as int? ?? 0) * 1000,
      ),
      fields: fields,
      display: metadata != null
          ? CredentialDisplay.fromJson(metadata['display']?[0] ?? {})
          : null,
      status: CredentialStatus.active,
      addedAt: DateTime.now(),
      subjectDid: credentialSubject['id']?.toString(),
    );

    // Save credential
    await _storageService.saveCredential(credential);

    // Log activity
    await _ref.read(activityProvider.notifier).addActivity(
      ActivityLog(
        id: const Uuid().v4(),
        type: ActivityType.credentialReceived,
        timestamp: DateTime.now(),
        credentialId: credential.id,
        credentialType: credential.credentialType,
        counterparty: credential.issuerName,
        description: '已取得 ${credential.credentialType}',
      ),
    );

    // Reload credentials
    await loadCredentials();

    return credential;
  }

  Future<void> deleteCredential(String id) async {
    final credential = _storageService.getCredential(id);
    await _storageService.deleteCredential(id);

    if (credential != null) {
      await _ref.read(activityProvider.notifier).addActivity(
        ActivityLog(
          id: const Uuid().v4(),
          type: ActivityType.credentialDeleted,
          timestamp: DateTime.now(),
          credentialId: id,
          credentialType: credential.credentialType,
          description: '已刪除 ${credential.credentialType}',
        ),
      );
    }

    await loadCredentials();
  }

  String _extractCredentialType(Map<String, dynamic>? vcData) {
    final types = vcData?['type'] as List<dynamic>?;
    if (types != null && types.length > 1) {
      return types.where((t) => t != 'VerifiableCredential').first.toString();
    }
    return 'Credential';
  }

  String _getFieldLabel(String key) {
    final labels = {
      'name': '姓名',
      'birthDate': '出生日期',
      'address': '地址',
      'licenseNumber': '證號',
      'issuanceDate': '發證日期',
      'expiryDate': '有效期限',
      'category': '類別',
      'phoneNumber': '電話號碼',
    };
    return labels[key] ?? key;
  }
}

final credentialsProvider =
    StateNotifierProvider<CredentialsNotifier, AsyncValue<List<StoredCredential>>>((ref) {
  return CredentialsNotifier(
    ref.watch(sdkServiceProvider),
    ref.watch(storageServiceProvider),
    ref,
  );
});

// ============ Activity Provider ============

class ActivityNotifier extends StateNotifier<List<ActivityLog>> {
  final StorageService _storageService;

  ActivityNotifier(this._storageService) : super([]);

  Future<void> loadActivities() async {
    state = _storageService.getAllActivities();
  }

  Future<void> addActivity(ActivityLog activity) async {
    await _storageService.addActivity(activity);
    state = [activity, ...state];
  }
}

final activityProvider =
    StateNotifierProvider<ActivityNotifier, List<ActivityLog>>((ref) {
  return ActivityNotifier(ref.watch(storageServiceProvider));
});

// ============ Settings Provider ============

class SettingsState {
  final bool biometricEnabled;
  final bool offlineMode;
  final DateTime? lastIssuerListUpdate;
  final DateTime? lastVCListUpdate;

  const SettingsState({
    this.biometricEnabled = false,
    this.offlineMode = false,
    this.lastIssuerListUpdate,
    this.lastVCListUpdate,
  });

  SettingsState copyWith({
    bool? biometricEnabled,
    bool? offlineMode,
    DateTime? lastIssuerListUpdate,
    DateTime? lastVCListUpdate,
  }) {
    return SettingsState(
      biometricEnabled: biometricEnabled ?? this.biometricEnabled,
      offlineMode: offlineMode ?? this.offlineMode,
      lastIssuerListUpdate: lastIssuerListUpdate ?? this.lastIssuerListUpdate,
      lastVCListUpdate: lastVCListUpdate ?? this.lastVCListUpdate,
    );
  }
}

class SettingsNotifier extends StateNotifier<SettingsState> {
  final StorageService _storageService;

  SettingsNotifier(this._storageService) : super(const SettingsState());

  Future<void> loadSettings() async {
    final biometricEnabled = await _storageService.isBiometricEnabled();
    state = state.copyWith(biometricEnabled: biometricEnabled);
  }

  Future<void> setBiometricEnabled(bool enabled) async {
    await _storageService.setBiometricEnabled(enabled);
    state = state.copyWith(biometricEnabled: enabled);
  }

  void setOfflineMode(bool enabled) {
    state = state.copyWith(offlineMode: enabled);
  }
}

final settingsProvider =
    StateNotifierProvider<SettingsNotifier, SettingsState>((ref) {
  return SettingsNotifier(ref.watch(storageServiceProvider));
});
