import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../providers/providers.dart';
import '../../router/app_router.dart';
import '../../widgets/credential_card.dart';
import '../../../core/theme/app_colors.dart';
import '../../../data/models/credential.dart';
import '../../../data/models/presentation_request.dart';

class SelectCredentialsScreen extends ConsumerStatefulWidget {
  final PresentationRequest? request;

  const SelectCredentialsScreen({super.key, this.request});

  @override
  ConsumerState<SelectCredentialsScreen> createState() =>
      _SelectCredentialsScreenState();
}

class _SelectCredentialsScreenState
    extends ConsumerState<SelectCredentialsScreen> {
  final Set<String> _selectedIds = {};

  @override
  Widget build(BuildContext context) {
    final credentialsAsync = ref.watch(credentialsProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('選擇憑證'),
      ),
      body: credentialsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('載入失敗: $e')),
        data: (credentials) {
          // Filter credentials that match the request
          final matchingCredentials = _filterMatchingCredentials(credentials);

          if (matchingCredentials.isEmpty) {
            return _buildNoMatchingCredentials();
          }

          return Column(
            children: [
              if (widget.request != null) _buildRequestInfo(),
              Expanded(
                child: ListView.builder(
                  padding: const EdgeInsets.all(16),
                  itemCount: matchingCredentials.length,
                  itemBuilder: (context, index) {
                    final credential = matchingCredentials[index];
                    final isSelected = _selectedIds.contains(credential.id);

                    return CredentialListTile(
                      credential: credential,
                      isSelected: isSelected,
                      onSelect: () {
                        setState(() {
                          if (isSelected) {
                            _selectedIds.remove(credential.id);
                          } else {
                            _selectedIds.add(credential.id);
                          }
                        });
                      },
                    );
                  },
                ),
              ),
              _buildBottomBar(credentials),
            ],
          );
        },
      ),
    );
  }

  List<StoredCredential> _filterMatchingCredentials(
      List<StoredCredential> credentials) {
    if (widget.request == null ||
        widget.request!.requestedCredentials.isEmpty) {
      return credentials.where((c) => c.isValid).toList();
    }

    // Match credentials by type
    final requestedTypes = widget.request!.requestedCredentials
        .map((r) => r.cardType.toLowerCase())
        .toSet();

    return credentials.where((c) {
      if (!c.isValid) return false;
      return requestedTypes.isEmpty ||
          requestedTypes.any((type) =>
              c.credentialType.toLowerCase().contains(type) ||
              type.contains(c.credentialType.toLowerCase()));
    }).toList();
  }

  Widget _buildRequestInfo() {
    return Container(
      margin: const EdgeInsets.all(16),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppColors.infoLight,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.business, color: AppColors.info, size: 20),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  widget.request!.verifierName ?? '驗證單位',
                  style: const TextStyle(
                    fontWeight: FontWeight.bold,
                    color: AppColors.info,
                  ),
                ),
              ),
            ],
          ),
          if (widget.request!.purpose != null) ...[
            const SizedBox(height: 8),
            Text(
              '目的: ${widget.request!.purpose}',
              style: const TextStyle(
                fontSize: 14,
                color: AppColors.textSecondary,
              ),
            ),
          ],
          const SizedBox(height: 8),
          Text(
            '請求的憑證: ${widget.request!.requestedCredentials.map((r) => r.name ?? r.cardType).join(', ')}',
            style: const TextStyle(
              fontSize: 14,
              color: AppColors.textSecondary,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildNoMatchingCredentials() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              width: 100,
              height: 100,
              decoration: BoxDecoration(
                color: AppColors.warningLight,
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.search_off,
                size: 48,
                color: AppColors.warning,
              ),
            ),
            const SizedBox(height: 24),
            const Text(
              '沒有符合的憑證',
              style: TextStyle(
                fontSize: 20,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 12),
            const Text(
              '您的皮夾中沒有符合此驗證請求的憑證',
              style: TextStyle(
                color: AppColors.textSecondary,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 32),
            OutlinedButton(
              onPressed: () => context.pop(),
              child: const Text('返回'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildBottomBar(List<StoredCredential> credentials) {
    final selectedCredentials = credentials
        .where((c) => _selectedIds.contains(c.id))
        .toList();

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
              '已選擇 ${_selectedIds.length} 張憑證',
              style: const TextStyle(
                color: AppColors.textSecondary,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 12),
            ElevatedButton(
              onPressed: _selectedIds.isEmpty
                  ? null
                  : () {
                      context.push(
                        AppRoutes.disclosureSelection,
                        extra: {
                          'request': widget.request,
                          'credentials': selectedCredentials,
                        },
                      );
                    },
              child: const Text('繼續'),
            ),
          ],
        ),
      ),
    );
  }
}
