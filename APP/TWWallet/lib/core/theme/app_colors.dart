import 'package:flutter/material.dart';

/// Government-style color palette for TW Digital Wallet
class AppColors {
  AppColors._();

  // Primary - Government Blue
  static const Color primary = Color(0xFF1565C0);
  static const Color primaryLight = Color(0xFF5E92F3);
  static const Color primaryDark = Color(0xFF003C8F);

  // Secondary - Trust Green
  static const Color secondary = Color(0xFF2E7D32);
  static const Color secondaryLight = Color(0xFF60AD5E);
  static const Color secondaryDark = Color(0xFF005005);

  // Accent - Digital Gold
  static const Color accent = Color(0xFFFFB300);
  static const Color accentLight = Color(0xFFFFE54C);
  static const Color accentDark = Color(0xFFC68400);

  // Neutral
  static const Color background = Color(0xFFF5F7FA);
  static const Color surface = Colors.white;
  static const Color surfaceVariant = Color(0xFFEEF2F6);
  static const Color textPrimary = Color(0xFF1A1A2E);
  static const Color textSecondary = Color(0xFF6B7280);
  static const Color textHint = Color(0xFF9CA3AF);
  static const Color divider = Color(0xFFE5E7EB);
  static const Color border = Color(0xFFD1D5DB);

  // Status
  static const Color success = Color(0xFF10B981);
  static const Color successLight = Color(0xFFD1FAE5);
  static const Color warning = Color(0xFFF59E0B);
  static const Color warningLight = Color(0xFFFEF3C7);
  static const Color error = Color(0xFFEF4444);
  static const Color errorLight = Color(0xFFFEE2E2);
  static const Color info = Color(0xFF3B82F6);
  static const Color infoLight = Color(0xFFDBEAFE);

  // Credential card gradients
  static const List<Color> cardGradientBlue = [
    Color(0xFF1565C0),
    Color(0xFF0D47A1),
  ];

  static const List<Color> cardGradientGreen = [
    Color(0xFF2E7D32),
    Color(0xFF1B5E20),
  ];

  static const List<Color> cardGradientPurple = [
    Color(0xFF7B1FA2),
    Color(0xFF4A148C),
  ];

  static const List<Color> cardGradientTeal = [
    Color(0xFF00897B),
    Color(0xFF004D40),
  ];

  // Shadows
  static const Color shadow = Color(0x1A000000);
  static const Color shadowDark = Color(0x33000000);
}
