import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';
import '../../providers/providers.dart';
import '../../../data/models/credential.dart';
import '../../../core/theme/app_colors.dart';
import '../../widgets/credential_card.dart';

class CredentialDetailScreen extends ConsumerWidget {
  final String credentialId;
  final StoredCredential? credential;

  const CredentialDetailScreen({
    super.key,
    required this.credentialId,
    this.credential,
  });

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final cred = credential ??
        ref.watch(credentialsProvider).whenData((list) {
          return list.firstWhere(
            (c) => c.id == credentialId,
            orElse: () => throw Exception('Credential not found'),
          );
        }).value;

    if (cred == null) {
      return Scaffold(
        appBar: AppBar(),
        body: const Center(child: CircularProgressIndicator()),
      );
    }

    return Scaffold(
      body: CustomScrollView(
        slivers: [
          SliverAppBar(
            expandedHeight: 280,
            pinned: true,
            flexibleSpace: FlexibleSpaceBar(
              background: Padding(
                padding: const EdgeInsets.only(top: 80),
                child: AspectRatio(
                  aspectRatio: 1.6,
                  child: CredentialCard(credential: cred),
                ),
              ),
            ),
            actions: [
              PopupMenuButton<String>(
                onSelected: (value) async {
                  if (value == 'delete') {
                    final confirmed = await _showDeleteDialog(context);
                    if (confirmed == true) {
                      await ref
                          .read(credentialsProvider.notifier)
                          .deleteCredential(cred.id);
                      if (context.mounted) {
                        context.pop();
                      }
                    }
                  }
                },
                itemBuilder: (context) => [
                  const PopupMenuItem(
                    value: 'verify',
                    child: ListTile(
                      leading: Icon(Icons.verified_user),
                      title: Text('驗證憑證'),
                      contentPadding: EdgeInsets.zero,
                    ),
                  ),
                  const PopupMenuItem(
                    value: 'transfer',
                    child: ListTile(
                      leading: Icon(Icons.send),
                      title: Text('轉移憑證'),
                      contentPadding: EdgeInsets.zero,
                    ),
                  ),
                  const PopupMenuDivider(),
                  const PopupMenuItem(
                    value: 'delete',
                    child: ListTile(
                      leading: Icon(Icons.delete, color: AppColors.error),
                      title: Text('刪除憑證',
                          style: TextStyle(color: AppColors.error)),
                      contentPadding: EdgeInsets.zero,
                    ),
                  ),
                ],
              ),
            ],
          ),
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildInfoSection(cred),
                  const SizedBox(height: 24),
                  _buildFieldsSection(cred),
                  const SizedBox(height: 24),
                  _buildMetadataSection(cred),
                  const SizedBox(height: 100),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildInfoSection(StoredCredential cred) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
        side: const BorderSide(color: AppColors.border),
      ),
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Icon(Icons.info_outline, color: AppColors.primary),
                const SizedBox(width: 8),
                const Text(
                  '憑證資訊',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            _buildInfoRow('類型', cred.credentialType),
            _buildInfoRow('發行機構', cred.issuerName),
            _buildInfoRow(
                '發證日期', DateFormat('yyyy/MM/dd').format(cred.issuedAt)),
            _buildInfoRow(
                '有效期限', DateFormat('yyyy/MM/dd').format(cred.expiresAt)),
            _buildInfoRow('狀態', _getStatusText(cred)),
          ],
        ),
      ),
    );
  }

  Widget _buildFieldsSection(StoredCredential cred) {
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
        side: const BorderSide(color: AppColors.border),
      ),
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Icon(Icons.list_alt, color: AppColors.primary),
                const SizedBox(width: 8),
                const Text(
                  '憑證欄位',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            ...cred.fields.map((field) => _buildFieldRow(field)),
          ],
        ),
      ),
    );
  }

  Widget _buildMetadataSection(StoredCredential cred) {
    return Card(
      elevation: 0,
      color: AppColors.surfaceVariant,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
      ),
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              '技術資訊',
              style: TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w600,
                color: AppColors.textSecondary,
              ),
            ),
            const SizedBox(height: 12),
            _buildTechRow('憑證 ID', cred.id),
            _buildTechRow('發行者 DID', cred.issuer),
            if (cred.subjectDid != null)
              _buildTechRow('持有者 DID', cred.subjectDid!),
            _buildTechRow(
                '新增時間', DateFormat('yyyy/MM/dd HH:mm').format(cred.addedAt)),
          ],
        ),
      ),
    );
  }

  Widget _buildInfoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 100,
            child: Text(
              label,
              style: const TextStyle(
                color: AppColors.textSecondary,
              ),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: const TextStyle(
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildFieldRow(CredentialField field) {
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 12),
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: AppColors.divider)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            flex: 2,
            child: Text(
              field.label,
              style: const TextStyle(
                color: AppColors.textSecondary,
              ),
            ),
          ),
          Expanded(
            flex: 3,
            child: Row(
              children: [
                Expanded(
                  child: Text(
                    field.value,
                    style: const TextStyle(
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ),
                if (field.isDisclosable)
                  const Icon(
                    Icons.visibility,
                    size: 16,
                    color: AppColors.textHint,
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTechRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            label,
            style: const TextStyle(
              fontSize: 12,
              color: AppColors.textHint,
            ),
          ),
          const SizedBox(height: 2),
          Text(
            value,
            style: const TextStyle(
              fontSize: 12,
              fontFamily: 'monospace',
            ),
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
        ],
      ),
    );
  }

  String _getStatusText(StoredCredential cred) {
    if (cred.isExpired) return '已過期';
    switch (cred.status) {
      case CredentialStatus.active:
        return '有效';
      case CredentialStatus.suspended:
        return '暫停使用';
      case CredentialStatus.revoked:
        return '已撤銷';
      case CredentialStatus.expired:
        return '已過期';
      default:
        return '未知';
    }
  }

  Future<bool?> _showDeleteDialog(BuildContext context) {
    return showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('刪除憑證'),
        content: const Text('確定要刪除此憑證嗎？此操作無法復原。'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('取消'),
          ),
          TextButton(
            onPressed: () => Navigator.of(context).pop(true),
            style: TextButton.styleFrom(foregroundColor: AppColors.error),
            child: const Text('刪除'),
          ),
        ],
      ),
    );
  }
}
