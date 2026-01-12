import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:uuid/uuid.dart';
import '../../../services/nfc_service.dart';
import '../../providers/providers.dart';
import '../../router/app_router.dart';
import '../../../core/theme/app_colors.dart';
import '../../../data/models/credential.dart';
import '../../../data/models/activity_log.dart';
import '../../widgets/credential_card.dart';

class NFCPresentScreen extends ConsumerStatefulWidget {
  final StoredCredential? credential;

  const NFCPresentScreen({super.key, this.credential});

  @override
  ConsumerState<NFCPresentScreen> createState() => _NFCPresentScreenState();
}

class _NFCPresentScreenState extends ConsumerState<NFCPresentScreen>
    with SingleTickerProviderStateMixin {
  final NFCService _nfcService = NFCService();

  NFCState _state = NFCState.selectCredential;
  StoredCredential? _selectedCredential;
  String? _errorMessage;
  bool _isNFCAvailable = false;

  late AnimationController _pulseController;
  late Animation<double> _pulseAnimation;

  @override
  void initState() {
    super.initState();
    _selectedCredential = widget.credential;
    if (_selectedCredential != null) {
      _state = NFCState.ready;
    }
    _checkNFCAvailability();

    _pulseController = AnimationController(
      duration: const Duration(milliseconds: 1500),
      vsync: this,
    )..repeat(reverse: true);

    _pulseAnimation = Tween<double>(begin: 1.0, end: 1.2).animate(
      CurvedAnimation(parent: _pulseController, curve: Curves.easeInOut),
    );
  }

  @override
  void dispose() {
    _pulseController.dispose();
    _nfcService.stopSession();
    super.dispose();
  }

  Future<void> _checkNFCAvailability() async {
    final available = await _nfcService.isNFCAvailable();
    setState(() {
      _isNFCAvailable = available;
      if (!available) {
        _errorMessage = '此裝置不支援 NFC 功能';
      }
    });
  }

  Future<void> _startNFCTransmission() async {
    if (_selectedCredential == null) return;

    setState(() {
      _state = NFCState.transmitting;
      _errorMessage = null;
    });

    try {
      final identityState = ref.read(identityProvider);
      if (identityState.didDocument == null) {
        throw Exception('找不到身分資料');
      }

      // Generate VP for NFC using SDK
      final sdkService = ref.read(sdkServiceProvider);
      final vpData = await sdkService.generateVPNFC(
        didDocument: identityState.didDocument!,
        vc: _selectedCredential!.rawToken,
      );

      // Start NFC write session
      await _nfcService.startWriteSession(
        vpData: vpData,
        onResult: (success, message) async {
          if (success) {
            // Log activity
            await ref.read(activityProvider.notifier).addActivity(
              ActivityLog(
                id: const Uuid().v4(),
                type: ActivityType.credentialPresented,
                timestamp: DateTime.now(),
                credentialId: _selectedCredential!.id,
                credentialType: _selectedCredential!.credentialType,
                counterparty: 'NFC 驗證',
                description: '已透過 NFC 出示 ${_selectedCredential!.credentialType}',
              ),
            );

            setState(() => _state = NFCState.success);
          } else {
            setState(() {
              _state = NFCState.error;
              _errorMessage = message;
            });
          }
        },
      );
    } catch (e) {
      setState(() {
        _state = NFCState.error;
        _errorMessage = e.toString();
      });
    }
  }

  void _selectCredential(StoredCredential credential) {
    setState(() {
      _selectedCredential = credential;
      _state = NFCState.ready;
    });
  }

  void _reset() {
    _nfcService.stopSession();
    setState(() {
      _state = _selectedCredential != null ? NFCState.ready : NFCState.selectCredential;
      _errorMessage = null;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('NFC 出示憑證'),
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    switch (_state) {
      case NFCState.selectCredential:
        return _buildCredentialSelection();
      case NFCState.ready:
        return _buildReadyState();
      case NFCState.transmitting:
        return _buildTransmittingState();
      case NFCState.success:
        return _buildSuccessState();
      case NFCState.error:
        return _buildErrorState();
    }
  }

  Widget _buildCredentialSelection() {
    final credentialsAsync = ref.watch(credentialsProvider);

    return credentialsAsync.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (e, _) => Center(child: Text('載入失敗: $e')),
      data: (credentials) {
        final validCredentials = credentials.where((c) => c.isValid).toList();

        if (validCredentials.isEmpty) {
          return _buildNoCredentials();
        }

        return Column(
          children: [
            Container(
              padding: const EdgeInsets.all(16),
              color: AppColors.infoLight,
              child: const Row(
                children: [
                  Icon(Icons.info_outline, color: AppColors.info),
                  SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      '選擇要透過 NFC 出示的憑證',
                      style: TextStyle(color: AppColors.info),
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              child: ListView.builder(
                padding: const EdgeInsets.all(16),
                itemCount: validCredentials.length,
                itemBuilder: (context, index) {
                  final credential = validCredentials[index];
                  return CredentialListTile(
                    credential: credential,
                    onTap: () => _selectCredential(credential),
                  );
                },
              ),
            ),
          ],
        );
      },
    );
  }

  Widget _buildNoCredentials() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.credit_card_off,
              size: 64,
              color: AppColors.textHint,
            ),
            const SizedBox(height: 24),
            const Text(
              '沒有可用的憑證',
              style: TextStyle(
                fontSize: 20,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 12),
            const Text(
              '請先取得憑證後再使用 NFC 出示功能',
              style: TextStyle(color: AppColors.textSecondary),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildReadyState() {
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          children: [
            if (_selectedCredential != null) ...[
              AspectRatio(
                aspectRatio: 1.6,
                child: CredentialCard(
                  credential: _selectedCredential!,
                  isCompact: true,
                ),
              ),
              const SizedBox(height: 24),
            ],
            Expanded(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Container(
                    width: 120,
                    height: 120,
                    decoration: BoxDecoration(
                      color: _isNFCAvailable
                          ? AppColors.primaryLight.withOpacity(0.2)
                          : AppColors.errorLight,
                      shape: BoxShape.circle,
                    ),
                    child: Icon(
                      Icons.nfc,
                      size: 64,
                      color: _isNFCAvailable ? AppColors.primary : AppColors.error,
                    ),
                  ),
                  const SizedBox(height: 32),
                  Text(
                    _isNFCAvailable ? '準備就緒' : 'NFC 不可用',
                    style: const TextStyle(
                      fontSize: 24,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 12),
                  Text(
                    _isNFCAvailable
                        ? '點擊下方按鈕開始 NFC 傳輸'
                        : '此裝置不支援 NFC 功能',
                    style: const TextStyle(
                      color: AppColors.textSecondary,
                    ),
                  ),
                ],
              ),
            ),
            Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                ElevatedButton.icon(
                  onPressed: _isNFCAvailable ? _startNFCTransmission : null,
                  icon: const Icon(Icons.nfc),
                  label: const Text('開始 NFC 傳輸'),
                ),
                const SizedBox(height: 12),
                TextButton(
                  onPressed: () {
                    setState(() {
                      _selectedCredential = null;
                      _state = NFCState.selectCredential;
                    });
                  },
                  child: const Text('選擇其他憑證'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTransmittingState() {
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            AnimatedBuilder(
              animation: _pulseAnimation,
              builder: (context, child) {
                return Transform.scale(
                  scale: _pulseAnimation.value,
                  child: Container(
                    width: 160,
                    height: 160,
                    decoration: BoxDecoration(
                      color: AppColors.primary.withOpacity(0.1),
                      shape: BoxShape.circle,
                    ),
                    child: Center(
                      child: Container(
                        width: 120,
                        height: 120,
                        decoration: BoxDecoration(
                          color: AppColors.primary.withOpacity(0.2),
                          shape: BoxShape.circle,
                        ),
                        child: const Icon(
                          Icons.nfc,
                          size: 64,
                          color: AppColors.primary,
                        ),
                      ),
                    ),
                  ),
                );
              },
            ),
            const SizedBox(height: 48),
            const Text(
              '請將裝置靠近 NFC 讀卡機',
              style: TextStyle(
                fontSize: 24,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 16),
            const Text(
              '保持裝置靠近直到傳輸完成',
              style: TextStyle(
                color: AppColors.textSecondary,
                fontSize: 16,
              ),
            ),
            const SizedBox(height: 48),
            const CircularProgressIndicator(),
            const SizedBox(height: 48),
            TextButton(
              onPressed: _reset,
              child: const Text('取消'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSuccessState() {
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              width: 120,
              height: 120,
              decoration: const BoxDecoration(
                color: AppColors.successLight,
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.check,
                size: 64,
                color: AppColors.success,
              ),
            ),
            const SizedBox(height: 32),
            const Text(
              'NFC 傳輸成功！',
              style: TextStyle(
                fontSize: 28,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 16),
            const Text(
              '您的憑證已成功傳輸給驗證方',
              style: TextStyle(
                color: AppColors.textSecondary,
                fontSize: 16,
              ),
            ),
            const Spacer(),
            Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                ElevatedButton(
                  onPressed: () => context.go(AppRoutes.home),
                  child: const Text('返回皮夾'),
                ),
                const SizedBox(height: 12),
                OutlinedButton(
                  onPressed: _reset,
                  child: const Text('再次傳輸'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildErrorState() {
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              width: 120,
              height: 120,
              decoration: const BoxDecoration(
                color: AppColors.errorLight,
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.error_outline,
                size: 64,
                color: AppColors.error,
              ),
            ),
            const SizedBox(height: 32),
            const Text(
              'NFC 傳輸失敗',
              style: TextStyle(
                fontSize: 28,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 16),
            if (_errorMessage != null)
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: AppColors.errorLight,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(
                  _errorMessage!,
                  style: const TextStyle(color: AppColors.error),
                  textAlign: TextAlign.center,
                ),
              ),
            const Spacer(),
            Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                ElevatedButton(
                  onPressed: _reset,
                  child: const Text('重試'),
                ),
                const SizedBox(height: 12),
                TextButton(
                  onPressed: () => context.pop(),
                  child: const Text('取消'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

enum NFCState {
  selectCredential,
  ready,
  transmitting,
  success,
  error,
}
