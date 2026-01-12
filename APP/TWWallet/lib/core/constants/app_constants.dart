/// Application-wide constants
class AppConstants {
  AppConstants._();

  // App info
  static const String appName = 'TW 數位皮夾';
  static const String appNameEn = 'TW Digital Wallet';
  static const String appVersion = '1.0.0';

  // API endpoints (configurable)
  static const String defaultFrontUrl = 'https://api.twdiw.gov.tw';

  // Storage keys
  static const String keyDIDDocument = 'did_document';
  static const String keyPublicKey = 'public_key';
  static const String keyCredentials = 'credentials';
  static const String keyIssuerList = 'issuer_list';
  static const String keyVCStatusList = 'vc_status_list';
  static const String keyActivityLog = 'activity_log';
  static const String keyOnboardingComplete = 'onboarding_complete';
  static const String keyBiometricEnabled = 'biometric_enabled';
  static const String keyPinHash = 'pin_hash';

  // Key generation
  static const String defaultKeyTag = 'tw_wallet_key';
  static const String defaultKeyType = 'P256';

  // Timeouts
  static const Duration apiTimeout = Duration(seconds: 30);
  static const Duration animationDuration = Duration(milliseconds: 300);
  static const Duration shortAnimationDuration = Duration(milliseconds: 150);

  // UI
  static const double cardAspectRatio = 1.586; // Credit card ratio
  static const double maxCardWidth = 380.0;
  static const double horizontalPadding = 20.0;
  static const double verticalPadding = 16.0;

  // Validation
  static const int pinLength = 6;
  static const int otpLength = 6;
}
