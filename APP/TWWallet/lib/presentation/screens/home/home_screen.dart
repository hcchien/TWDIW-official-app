import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../providers/providers.dart';
import '../../router/app_router.dart';
import '../../widgets/credential_card.dart';
import '../../../core/theme/app_colors.dart';

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  int _selectedIndex = 0;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: IndexedStack(
        index: _selectedIndex,
        children: const [
          _WalletView(),
          _ScanView(),
          _HistoryView(),
          _SettingsView(),
        ],
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: _selectedIndex,
        onDestinationSelected: (index) {
          setState(() {
            _selectedIndex = index;
          });
        },
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.wallet_outlined),
            selectedIcon: Icon(Icons.wallet),
            label: '皮夾',
          ),
          NavigationDestination(
            icon: Icon(Icons.qr_code_scanner_outlined),
            selectedIcon: Icon(Icons.qr_code_scanner),
            label: '掃描',
          ),
          NavigationDestination(
            icon: Icon(Icons.history_outlined),
            selectedIcon: Icon(Icons.history),
            label: '紀錄',
          ),
          NavigationDestination(
            icon: Icon(Icons.settings_outlined),
            selectedIcon: Icon(Icons.settings),
            label: '設定',
          ),
        ],
      ),
    );
  }
}

class _WalletView extends ConsumerWidget {
  const _WalletView();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final credentialsAsync = ref.watch(credentialsProvider);

    return SafeArea(
      child: CustomScrollView(
        slivers: [
          SliverAppBar(
            floating: true,
            title: const Text('我的皮夾'),
            actions: [
              IconButton(
                icon: const Icon(Icons.add),
                onPressed: () => context.push(AppRoutes.scanReceive),
                tooltip: '新增憑證',
              ),
            ],
          ),
          credentialsAsync.when(
            loading: () => const SliverFillRemaining(
              child: Center(child: CircularProgressIndicator()),
            ),
            error: (error, _) => SliverFillRemaining(
              child: Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Icon(Icons.error_outline, size: 64, color: AppColors.error),
                    const SizedBox(height: 16),
                    Text('載入失敗: $error'),
                    const SizedBox(height: 16),
                    ElevatedButton(
                      onPressed: () =>
                          ref.read(credentialsProvider.notifier).loadCredentials(),
                      child: const Text('重試'),
                    ),
                  ],
                ),
              ),
            ),
            data: (credentials) {
              if (credentials.isEmpty) {
                return SliverFillRemaining(
                  child: _buildEmptyState(context),
                );
              }
              return SliverList(
                delegate: SliverChildBuilderDelegate(
                  (context, index) {
                    final credential = credentials[index];
                    return Padding(
                      padding: EdgeInsets.only(
                        top: index == 0 ? 16 : 0,
                        bottom: index == credentials.length - 1 ? 100 : 0,
                      ),
                      child: AspectRatio(
                        aspectRatio: 1.6,
                        child: CredentialCard(
                          credential: credential,
                          onTap: () => context.push(
                            '/credential/${credential.id}',
                            extra: credential,
                          ),
                        ),
                      ),
                    );
                  },
                  childCount: credentials.length,
                ),
              );
            },
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyState(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              width: 120,
              height: 120,
              decoration: BoxDecoration(
                color: AppColors.primaryLight.withOpacity(0.2),
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.wallet,
                size: 64,
                color: AppColors.primary,
              ),
            ),
            const SizedBox(height: 32),
            const Text(
              '您的皮夾是空的',
              style: TextStyle(
                fontSize: 24,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 12),
            const Text(
              '掃描 QR Code 來取得您的第一張數位憑證',
              style: TextStyle(
                fontSize: 14,
                color: AppColors.textSecondary,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 32),
            ElevatedButton.icon(
              onPressed: () => context.push(AppRoutes.scanReceive),
              icon: const Icon(Icons.qr_code_scanner),
              label: const Text('掃描取得憑證'),
            ),
          ],
        ),
      ),
    );
  }
}

class _ScanView extends StatelessWidget {
  const _ScanView();

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const SizedBox(height: 24),
            const Text(
              '掃描與傳輸',
              style: TextStyle(
                fontSize: 28,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 8),
            const Text(
              '選擇您要執行的操作',
              style: TextStyle(
                color: AppColors.textSecondary,
              ),
            ),
            const SizedBox(height: 24),
            // QR Code Section
            const Text(
              'QR Code',
              style: TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w600,
                color: AppColors.primary,
              ),
            ),
            const SizedBox(height: 12),
            _buildActionCard(
              context,
              icon: Icons.download,
              title: '取得憑證',
              description: '掃描發行單位的 QR Code\n取得新的數位憑證',
              onTap: () => context.push(AppRoutes.scanReceive),
            ),
            const SizedBox(height: 12),
            _buildActionCard(
              context,
              icon: Icons.upload,
              title: '出示憑證',
              description: '掃描驗證單位的 QR Code\n選擇要出示的憑證資訊',
              onTap: () => context.push(AppRoutes.scanPresent),
            ),
            const SizedBox(height: 24),
            // NFC Section
            const Text(
              'NFC 傳輸',
              style: TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w600,
                color: AppColors.secondary,
              ),
            ),
            const SizedBox(height: 12),
            _buildActionCard(
              context,
              icon: Icons.nfc,
              title: 'NFC 出示憑證',
              description: '透過 NFC 將憑證傳輸\n給驗證方的讀卡機',
              onTap: () => context.push(AppRoutes.nfcPresent),
              color: AppColors.secondary,
            ),
            const SizedBox(height: 12),
            _buildActionCard(
              context,
              icon: Icons.nfc,
              title: 'NFC 驗證憑證',
              description: '讀取他人的 NFC 標籤\n驗證其憑證有效性',
              onTap: () => context.push(AppRoutes.nfcVerify),
              color: AppColors.secondary,
            ),
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }

  Widget _buildActionCard(
    BuildContext context, {
    required IconData icon,
    required String title,
    required String description,
    required VoidCallback onTap,
    Color color = AppColors.primary,
  }) {
    final lightColor = color == AppColors.secondary
        ? AppColors.secondaryLight
        : AppColors.primaryLight;
    return Card(
      elevation: 0,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
        side: const BorderSide(color: AppColors.border),
      ),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(16),
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Row(
            children: [
              Container(
                width: 64,
                height: 64,
                decoration: BoxDecoration(
                  color: lightColor.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(16),
                ),
                child: Icon(
                  icon,
                  size: 32,
                  color: color,
                ),
              ),
              const SizedBox(width: 20),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      title,
                      style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      description,
                      style: const TextStyle(
                        fontSize: 14,
                        color: AppColors.textSecondary,
                        height: 1.4,
                      ),
                    ),
                  ],
                ),
              ),
              const Icon(
                Icons.chevron_right,
                color: AppColors.textHint,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _HistoryView extends ConsumerWidget {
  const _HistoryView();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final activities = ref.watch(activityProvider);

    return SafeArea(
      child: CustomScrollView(
        slivers: [
          const SliverAppBar(
            floating: true,
            title: Text('活動紀錄'),
          ),
          if (activities.isEmpty)
            const SliverFillRemaining(
              child: Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(Icons.history, size: 64, color: AppColors.textHint),
                    SizedBox(height: 16),
                    Text(
                      '暫無活動紀錄',
                      style: TextStyle(color: AppColors.textSecondary),
                    ),
                  ],
                ),
              ),
            )
          else
            SliverList(
              delegate: SliverChildBuilderDelegate(
                (context, index) {
                  final activity = activities[index];
                  return ListTile(
                    leading: CircleAvatar(
                      backgroundColor: activity.success
                          ? AppColors.successLight
                          : AppColors.errorLight,
                      child: Text(activity.typeIcon),
                    ),
                    title: Text(activity.typeDisplayName),
                    subtitle: Text(
                      activity.description,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    trailing: Text(
                      _formatTime(activity.timestamp),
                      style: const TextStyle(
                        fontSize: 12,
                        color: AppColors.textHint,
                      ),
                    ),
                  );
                },
                childCount: activities.length,
              ),
            ),
        ],
      ),
    );
  }

  String _formatTime(DateTime time) {
    final now = DateTime.now();
    final diff = now.difference(time);

    if (diff.inMinutes < 1) return '剛剛';
    if (diff.inHours < 1) return '${diff.inMinutes} 分鐘前';
    if (diff.inDays < 1) return '${diff.inHours} 小時前';
    if (diff.inDays < 7) return '${diff.inDays} 天前';
    return '${time.month}/${time.day}';
  }
}

class _SettingsView extends ConsumerWidget {
  const _SettingsView();

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final settings = ref.watch(settingsProvider);

    return SafeArea(
      child: ListView(
        children: [
          const SizedBox(height: 16),
          const Padding(
            padding: EdgeInsets.symmetric(horizontal: 20),
            child: Text(
              '設定',
              style: TextStyle(
                fontSize: 28,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
          const SizedBox(height: 24),
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
                trailing: const Icon(Icons.chevron_right),
                onTap: () {
                  // TODO: Download issuer list
                },
              ),
            ],
          ),
          _buildSection(
            title: '關於',
            children: [
              const ListTile(
                title: Text('版本'),
                trailing: Text('1.0.0'),
              ),
              ListTile(
                title: const Text('隱私權政策'),
                trailing: const Icon(Icons.chevron_right),
                onTap: () {},
              ),
              ListTile(
                title: const Text('服務條款'),
                trailing: const Icon(Icons.chevron_right),
                onTap: () {},
              ),
            ],
          ),
          const SizedBox(height: 32),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 20),
            child: OutlinedButton(
              onPressed: () {
                // TODO: Reset wallet
              },
              style: OutlinedButton.styleFrom(
                foregroundColor: AppColors.error,
                side: const BorderSide(color: AppColors.error),
              ),
              child: const Text('重置皮夾'),
            ),
          ),
          const SizedBox(height: 100),
        ],
      ),
    );
  }

  Widget _buildSection({required String title, required List<Widget> children}) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
          child: Text(
            title,
            style: const TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.w600,
              color: AppColors.textSecondary,
            ),
          ),
        ),
        ...children,
        const SizedBox(height: 16),
      ],
    );
  }
}
