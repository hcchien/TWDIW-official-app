import 'package:flutter/material.dart';
import '../../data/models/credential.dart';
import '../../core/theme/app_colors.dart';
import 'package:intl/intl.dart';

class CredentialCard extends StatelessWidget {
  final StoredCredential credential;
  final VoidCallback? onTap;
  final bool isCompact;

  const CredentialCard({
    super.key,
    required this.credential,
    this.onTap,
    this.isCompact = false,
  });

  List<Color> get _gradientColors {
    final type = credential.credentialType.toLowerCase();
    if (type.contains('driver') || type.contains('駕照')) {
      return AppColors.cardGradientBlue;
    } else if (type.contains('student') || type.contains('學生')) {
      return AppColors.cardGradientGreen;
    } else if (type.contains('telecom') || type.contains('電信')) {
      return AppColors.cardGradientPurple;
    }
    return AppColors.cardGradientTeal;
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        margin: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
        decoration: BoxDecoration(
          gradient: LinearGradient(
            colors: _gradientColors,
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
          ),
          borderRadius: BorderRadius.circular(20),
          boxShadow: [
            BoxShadow(
              color: _gradientColors.first.withOpacity(0.4),
              blurRadius: 20,
              offset: const Offset(0, 10),
            ),
          ],
        ),
        child: ClipRRect(
          borderRadius: BorderRadius.circular(20),
          child: Stack(
            children: [
              // Background pattern
              Positioned(
                right: -50,
                top: -50,
                child: Container(
                  width: 200,
                  height: 200,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: Colors.white.withOpacity(0.1),
                  ),
                ),
              ),
              Positioned(
                right: -20,
                bottom: -60,
                child: Container(
                  width: 150,
                  height: 150,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: Colors.white.withOpacity(0.05),
                  ),
                ),
              ),
              // Content
              Padding(
                padding: EdgeInsets.all(isCompact ? 16 : 24),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: Colors.white.withOpacity(0.2),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: const Icon(
                            Icons.badge,
                            color: Colors.white,
                            size: 24,
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                credential.credentialType,
                                style: const TextStyle(
                                  color: Colors.white,
                                  fontSize: 18,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                              Text(
                                credential.issuerName,
                                style: TextStyle(
                                  color: Colors.white.withOpacity(0.8),
                                  fontSize: 12,
                                ),
                              ),
                            ],
                          ),
                        ),
                        _buildStatusBadge(),
                      ],
                    ),
                    if (!isCompact) ...[
                      const Spacer(),
                      // Display first 2 fields
                      ...credential.fields.take(2).map((field) => Padding(
                            padding: const EdgeInsets.only(bottom: 4),
                            child: Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(
                                  field.label,
                                  style: TextStyle(
                                    color: Colors.white.withOpacity(0.7),
                                    fontSize: 12,
                                  ),
                                ),
                                Text(
                                  field.value,
                                  style: const TextStyle(
                                    color: Colors.white,
                                    fontSize: 14,
                                    fontWeight: FontWeight.w500,
                                  ),
                                ),
                              ],
                            ),
                          )),
                      const SizedBox(height: 12),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          _buildDateInfo(
                            '發證日期',
                            DateFormat('yyyy/MM/dd').format(credential.issuedAt),
                          ),
                          _buildDateInfo(
                            '有效期限',
                            DateFormat('yyyy/MM/dd').format(credential.expiresAt),
                          ),
                        ],
                      ),
                    ],
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildStatusBadge() {
    Color color;
    String text;
    IconData icon;

    switch (credential.status) {
      case CredentialStatus.active:
        if (credential.isExpired) {
          color = AppColors.warning;
          text = '已過期';
          icon = Icons.schedule;
        } else {
          color = AppColors.success;
          text = '有效';
          icon = Icons.check_circle;
        }
        break;
      case CredentialStatus.suspended:
        color = AppColors.warning;
        text = '暫停';
        icon = Icons.pause_circle;
        break;
      case CredentialStatus.revoked:
        color = AppColors.error;
        text = '已撤銷';
        icon = Icons.cancel;
        break;
      case CredentialStatus.expired:
        color = AppColors.warning;
        text = '已過期';
        icon = Icons.schedule;
        break;
      default:
        color = AppColors.textHint;
        text = '未知';
        icon = Icons.help;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withOpacity(0.2),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withOpacity(0.5)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, color: color, size: 14),
          const SizedBox(width: 4),
          Text(
            text,
            style: TextStyle(
              color: color,
              fontSize: 12,
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDateInfo(String label, String value) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: TextStyle(
            color: Colors.white.withOpacity(0.6),
            fontSize: 10,
          ),
        ),
        Text(
          value,
          style: const TextStyle(
            color: Colors.white,
            fontSize: 12,
            fontWeight: FontWeight.w500,
          ),
        ),
      ],
    );
  }
}

/// Compact credential card for lists
class CredentialListTile extends StatelessWidget {
  final StoredCredential credential;
  final VoidCallback? onTap;
  final bool isSelected;
  final VoidCallback? onSelect;

  const CredentialListTile({
    super.key,
    required this.credential,
    this.onTap,
    this.isSelected = false,
    this.onSelect,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: isSelected
            ? const BorderSide(color: AppColors.primary, width: 2)
            : BorderSide.none,
      ),
      child: InkWell(
        onTap: onSelect ?? onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              if (onSelect != null) ...[
                Checkbox(
                  value: isSelected,
                  onChanged: (_) => onSelect?.call(),
                  activeColor: AppColors.primary,
                ),
                const SizedBox(width: 8),
              ],
              Container(
                width: 48,
                height: 48,
                decoration: BoxDecoration(
                  color: AppColors.primaryLight.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Icon(
                  Icons.badge,
                  color: AppColors.primary,
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      credential.credentialType,
                      style: const TextStyle(
                        fontWeight: FontWeight.w600,
                        fontSize: 16,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      credential.issuerName,
                      style: const TextStyle(
                        color: AppColors.textSecondary,
                        fontSize: 12,
                      ),
                    ),
                  ],
                ),
              ),
              Icon(
                credential.isValid ? Icons.check_circle : Icons.warning,
                color: credential.isValid ? AppColors.success : AppColors.warning,
              ),
            ],
          ),
        ),
      ),
    );
  }
}
