import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:uuid/uuid.dart';
import '../../providers/providers.dart';
import '../../router/app_router.dart';
import '../../../core/theme/app_colors.dart';
import '../../../data/models/credential.dart';
import '../../../data/models/presentation_request.dart';
import '../../../data/models/activity_log.dart';

class DisclosureSelectionScreen extends ConsumerStatefulWidget {
  final PresentationRequest? request;
  final List<StoredCredential> selectedCredentials;

  const DisclosureSelectionScreen({
    super.key,
    this.request,
    required this.selectedCredentials,
  });

  @override
  ConsumerState<DisclosureSelectionScreen> createState() =>
      _DisclosureSelectionScreenState();
}

class _DisclosureSelectionScreenState
    extends ConsumerState<DisclosureSelectionScreen> {
  // Map of credentialId -> Set of selected field keys
  late Map<String, Set<String>> _selectedFields;
  bool _isSubmitting = false;

  @override
  void initState() {
    super.initState();
    _initializeSelectedFields();
  }

  void _initializeSelectedFields() {
    _selectedFields = {};
    for (final credential in widget.selectedCredentials) {
      final requiredFields = _getRequiredFields(credential);
      _selectedFields[credential.id] = requiredFields.toSet();
    }
  }

  List<String> _getRequiredFields(StoredCredential credential) {
    if (widget.request == null) return [];

    for (final req in widget.request!.requestedCredentials) {
      if (credential.credentialType.toLowerCase().contains(req.cardType.toLowerCase()) ||
          req.cardType.toLowerCase().contains(credential.credentialType.toLowerCase())) {
        return req.requiredFields;
      }
    }
    return [];
  }

  List<String> _getOptionalFields(StoredCredential credential) {
    if (widget.request == null) return credential.fields.map((f) => f.key).toList();

    for (final req in widget.request!.requestedCredentials) {
      if (credential.credentialType.toLowerCase().contains(req.cardType.toLowerCase()) ||
          req.cardType.toLowerCase().contains(credential.credentialType.toLowerCase())) {
        return req.optionalFields;
      }
    }
    return [];
  }

  bool _isFieldRequired(StoredCredential credential, String fieldKey) {
    return _getRequiredFields(credential).contains(fieldKey);
  }

  Future<void> _submitPresentation() async {
    setState(() => _isSubmitting = true);

    try {
      final identityState = ref.read(identityProvider);
      if (identityState.didDocument == null) {
        throw Exception('No identity found');
      }

      final sdkService = ref.read(sdkServiceProvider);

      // Prepare VCs with selected fields
      final vcs = <Map<String, dynamic>>[];
      for (final credential in widget.selectedCredentials) {
        final selectedFieldKeys = _selectedFields[credential.id] ?? {};
        vcs.add({
          'credential': credential.rawToken,
          'fields': selectedFieldKeys.toList(),
        });
      }

      // Generate VP
      await sdkService.generateVP(
        didDocument: identityState.didDocument!,
        requestToken: widget.request?.requestToken ?? '',
        vcs: vcs,
      );

      // Log activity
      for (final credential in widget.selectedCredentials) {
        await ref.read(activityProvider.notifier).addActivity(
              ActivityLog(
                id: const Uuid().v4(),
                type: ActivityType.credentialPresented,
                timestamp: DateTime.now(),
                credentialId: credential.id,
                credentialType: credential.credentialType,
                counterparty: widget.request?.verifierName ?? '驗證單位',
                description:
                    '已出示 ${credential.credentialType} 給 ${widget.request?.verifierName ?? "驗證單位"}',
              ),
            );
      }

      if (mounted) {
        context.go(AppRoutes.presentationSuccess);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('出示憑證失敗: $e'),
            backgroundColor: AppColors.error,
          ),
        );
        setState(() => _isSubmitting = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('選擇揭露欄位'),
      ),
      body: Column(
        children: [
          _buildWarningBanner(),
          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.all(16),
              itemCount: widget.selectedCredentials.length,
              itemBuilder: (context, index) {
                final credential = widget.selectedCredentials[index];
                return _buildCredentialCard(credential);
              },
            ),
          ),
          _buildBottomBar(),
        ],
      ),
    );
  }

  Widget _buildWarningBanner() {
    return Container(
      margin: const EdgeInsets.all(16),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppColors.warningLight,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Row(
        children: [
          const Icon(Icons.info_outline, color: AppColors.warning),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  '選擇性揭露',
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    color: AppColors.warning,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  '您可以選擇要分享哪些資訊。必要欄位無法取消選取。',
                  style: TextStyle(
                    fontSize: 12,
                    color: AppColors.warning.withOpacity(0.8),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCredentialCard(StoredCredential credential) {
    final selectedKeys = _selectedFields[credential.id] ?? {};

    return Card(
      margin: const EdgeInsets.only(bottom: 16),
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.all(16),
            decoration: const BoxDecoration(
              color: AppColors.primaryLight,
              borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
            ),
            child: Row(
              children: [
                const Icon(Icons.badge, color: Colors.white),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        credential.credentialType,
                        style: const TextStyle(
                          color: Colors.white,
                          fontWeight: FontWeight.bold,
                          fontSize: 16,
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
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              children: credential.fields.map((field) {
                final isRequired = _isFieldRequired(credential, field.key);
                final isSelected = selectedKeys.contains(field.key);

                return CheckboxListTile(
                  title: Text(field.label),
                  subtitle: Text(
                    field.value,
                    style: const TextStyle(
                      color: AppColors.textSecondary,
                      fontSize: 12,
                    ),
                  ),
                  secondary: isRequired
                      ? Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 4,
                          ),
                          decoration: BoxDecoration(
                            color: AppColors.errorLight,
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: const Text(
                            '必要',
                            style: TextStyle(
                              fontSize: 10,
                              color: AppColors.error,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        )
                      : null,
                  value: isSelected,
                  onChanged: isRequired
                      ? null
                      : (value) {
                          setState(() {
                            if (value == true) {
                              selectedKeys.add(field.key);
                            } else {
                              selectedKeys.remove(field.key);
                            }
                          });
                        },
                  activeColor: AppColors.primary,
                  contentPadding: EdgeInsets.zero,
                );
              }).toList(),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildBottomBar() {
    final totalFields = widget.selectedCredentials
        .fold<int>(0, (sum, c) => sum + (_selectedFields[c.id]?.length ?? 0));

    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: AppColors.surface,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.1),
            blurRadius: 10,
            offset: const Offset(0, -4),
          ),
        ],
      ),
      child: SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              '將分享 $totalFields 個欄位',
              style: const TextStyle(
                color: AppColors.textSecondary,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 12),
            ElevatedButton(
              onPressed: _isSubmitting ? null : _submitPresentation,
              child: _isSubmitting
                  ? const SizedBox(
                      width: 24,
                      height: 24,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        color: Colors.white,
                      ),
                    )
                  : const Text('確認出示'),
            ),
            const SizedBox(height: 8),
            TextButton(
              onPressed: _isSubmitting ? null : () => context.pop(),
              child: const Text('取消'),
            ),
          ],
        ),
      ),
    );
  }
}
