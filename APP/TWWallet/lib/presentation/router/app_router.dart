import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../screens/splash/splash_screen.dart';
import '../screens/onboarding/onboarding_screen.dart';
import '../screens/home/home_screen.dart';
import '../screens/credentials/credential_detail_screen.dart';
import '../screens/receive/scan_qr_screen.dart';
import '../screens/receive/otp_input_screen.dart';
import '../screens/receive/receive_success_screen.dart';
import '../screens/present/scan_verifier_screen.dart';
import '../screens/present/select_credentials_screen.dart';
import '../screens/present/disclosure_selection_screen.dart';
import '../screens/present/presentation_success_screen.dart';
import '../screens/nfc/nfc_present_screen.dart';
import '../screens/nfc/nfc_verify_screen.dart';
import '../screens/history/activity_log_screen.dart';
import '../screens/settings/settings_screen.dart';
import '../../data/models/credential.dart';
import '../../data/models/presentation_request.dart';

/// Route names
class AppRoutes {
  static const String splash = '/';
  static const String onboarding = '/onboarding';
  static const String home = '/home';
  static const String credentialDetail = '/credential/:id';
  static const String scanReceive = '/receive/scan';
  static const String otpInput = '/receive/otp';
  static const String receiveSuccess = '/receive/success';
  static const String scanPresent = '/present/scan';
  static const String selectCredentials = '/present/select';
  static const String disclosureSelection = '/present/disclosure';
  static const String presentationSuccess = '/present/success';
  static const String nfcPresent = '/nfc/present';
  static const String nfcVerify = '/nfc/verify';
  static const String activityLog = '/history';
  static const String settings = '/settings';
}

final routerProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: AppRoutes.splash,
    debugLogDiagnostics: true,
    routes: [
      GoRoute(
        path: AppRoutes.splash,
        builder: (context, state) => const SplashScreen(),
      ),
      GoRoute(
        path: AppRoutes.onboarding,
        builder: (context, state) => const OnboardingScreen(),
      ),
      GoRoute(
        path: AppRoutes.home,
        builder: (context, state) => const HomeScreen(),
      ),
      GoRoute(
        path: '/credential/:id',
        builder: (context, state) {
          final credential = state.extra as StoredCredential?;
          return CredentialDetailScreen(
            credentialId: state.pathParameters['id']!,
            credential: credential,
          );
        },
      ),
      GoRoute(
        path: AppRoutes.scanReceive,
        builder: (context, state) => const ScanQRScreen(),
      ),
      GoRoute(
        path: AppRoutes.otpInput,
        builder: (context, state) {
          final qrData = state.extra as String?;
          return OTPInputScreen(qrData: qrData ?? '');
        },
      ),
      GoRoute(
        path: AppRoutes.receiveSuccess,
        builder: (context, state) {
          final credential = state.extra as StoredCredential?;
          return ReceiveSuccessScreen(credential: credential);
        },
      ),
      GoRoute(
        path: AppRoutes.scanPresent,
        builder: (context, state) => const ScanVerifierScreen(),
      ),
      GoRoute(
        path: AppRoutes.selectCredentials,
        builder: (context, state) {
          final request = state.extra as PresentationRequest?;
          return SelectCredentialsScreen(request: request);
        },
      ),
      GoRoute(
        path: AppRoutes.disclosureSelection,
        builder: (context, state) {
          final data = state.extra as Map<String, dynamic>?;
          return DisclosureSelectionScreen(
            request: data?['request'] as PresentationRequest?,
            selectedCredentials: data?['credentials'] as List<StoredCredential>? ?? [],
          );
        },
      ),
      GoRoute(
        path: AppRoutes.presentationSuccess,
        builder: (context, state) => const PresentationSuccessScreen(),
      ),
      GoRoute(
        path: AppRoutes.nfcPresent,
        builder: (context, state) {
          final credential = state.extra as StoredCredential?;
          return NFCPresentScreen(credential: credential);
        },
      ),
      GoRoute(
        path: AppRoutes.nfcVerify,
        builder: (context, state) => const NFCVerifyScreen(),
      ),
      GoRoute(
        path: AppRoutes.activityLog,
        builder: (context, state) => const ActivityLogScreen(),
      ),
      GoRoute(
        path: AppRoutes.settings,
        builder: (context, state) => const SettingsScreen(),
      ),
    ],
    errorBuilder: (context, state) => Scaffold(
      body: Center(
        child: Text('Page not found: ${state.uri}'),
      ),
    ),
  );
});
