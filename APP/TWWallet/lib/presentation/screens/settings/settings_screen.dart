import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../providers/providers.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/constants/app_constants.dart';

class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final settings = ref.watch(settingsProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('設定'),
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          _buildSection(
            title: '安全性',
            children: [
              SwitchListTile(
                title: const Text('生物辨識解鎖'),
                subtitle: const Text('使用 Face ID 或指紋解鎖'),
                value: settings.biometricEnabled,
                onChanged: (value) {
                  ref.read(settingsProvider.notifier).setBiometricEnabled(value);
                },
              ),
              ListTile(
                title: const Text('變更 PIN 碼'),
                trailing: const Icon(Icons.chevron_right),
                onTap: () {
                  // TODO: Change PIN
                },
              ),
            ],
          ),
          _buildSection(
            title: '離線模式',
            children: [
              SwitchListTile(
                title: const Text('離線驗證'),
                subtitle: const Text('允許在無網路時驗證憑證'),
                value: settings.offlineMode,
                onChanged: (value) {
                  ref.read(settingsProvider.notifier).setOfflineMode(value);
                },
              ),
              ListTile(
                title: const Text('更新發行機構清單'),
                subtitle: Text(
                  settings.lastIssuerListUpdate != null
                      ? '上次更新: ${settings.lastIssuerListUpdate}'
                      : '尚未更新',
                ),
                trailing: const Icon(Icons.download),
                onTap: () async {
                  // TODO: Download issuer list
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('正在更新發行機構清單...')),
                  );
                },
              ),
              ListTile(
                title: const Text('更新憑證狀態清單'),
                subtitle: Text(
                  settings.lastVCListUpdate != null
                      ? '上次更新: ${settings.lastVCListUpdate}'
                      : '尚未更新',
                ),
                trailing: const Icon(Icons.download),
                onTap: () async {
                  // TODO: Download VC status list
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('正在更新憑證狀態清單...')),
                  );
                },
              ),
            ],
          ),
          _buildSection(
            title: '關於',
            children: [
              ListTile(
                title: const Text('版本'),
                trailing: Text(
                  AppConstants.appVersion,
                  style: const TextStyle(color: AppColors.textSecondary),
                ),
              ),
              ListTile(
                title: const Text('SDK 版本'),
                trailing: Text(
                  ref.read(sdkServiceProvider).getVersion(),
                  style: const TextStyle(color: AppColors.textSecondary),
                ),
              ),
              ListTile(
                title: const Text('隱私權政策'),
                trailing: const Icon(Icons.open_in_new),
                onTap: () {
                  // TODO: Open privacy policy
                },
              ),
              ListTile(
                title: const Text('服務條款'),
                trailing: const Icon(Icons.open_in_new),
                onTap: () {
                  // TODO: Open terms of service
                },
              ),
              ListTile(
                title: const Text('開放原始碼授權'),
                trailing: const Icon(Icons.chevron_right),
                onTap: () {
                  showLicensePage(
                    context: context,
                    applicationName: AppConstants.appName,
                    applicationVersion: AppConstants.appVersion,
                  );
                },
              ),
            ],
          ),
          const SizedBox(height: 32),
          OutlinedButton(
            onPressed: () => _showResetDialog(context, ref),
            style: OutlinedButton.styleFrom(
              foregroundColor: AppColors.error,
              side: const BorderSide(color: AppColors.error),
            ),
            child: const Text('重置皮夾'),
          ),
          const SizedBox(height: 48),
        ],
      ),
    );
  }

  Widget _buildSection({
    required String title,
    required List<Widget> children,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          child: Text(
            title,
            style: const TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.w600,
              color: AppColors.primary,
            ),
          ),
        ),
        Card(
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
          child: Column(children: children),
        ),
        const SizedBox(height: 24),
      ],
    );
  }

  Future<void> _showResetDialog(BuildContext context, WidgetRef ref) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('重置皮夾'),
        content: const Text(
          '此操作將刪除所有儲存的憑證和身分資料。\n\n此操作無法復原，您確定要繼續嗎？',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('取消'),
          ),
          TextButton(
            onPressed: () => Navigator.of(context).pop(true),
            style: TextButton.styleFrom(foregroundColor: AppColors.error),
            child: const Text('重置'),
          ),
        ],
      ),
    );

    if (confirmed == true && context.mounted) {
      // TODO: Reset wallet
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('皮夾已重置')),
      );
    }
  }
}
