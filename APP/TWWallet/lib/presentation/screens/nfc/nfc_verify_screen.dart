import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../../services/nfc_service.dart';
import '../../providers/providers.dart';
import '../../router/app_router.dart';
import '../../../core/theme/app_colors.dart';

class NFCVerifyScreen extends ConsumerStatefulWidget {
  const NFCVerifyScreen({super.key});

  @override
  ConsumerState<NFCVerifyScreen> createState() => _NFCVerifyScreenState();
}

class _NFCVerifyScreenState extends ConsumerState<NFCVerifyScreen>
    with SingleTickerProviderStateMixin {
  final NFCService _nfcService = NFCService();

  NFCVerifyState _state = NFCVerifyState.idle;
  String? _errorMessage;
  Map<String, dynamic>? _verificationResult;
  bool _isNFCAvailable = false;

  late AnimationController _pulseController;
  late Animation<double> _pulseAnimation;

  @override
  void initState() {
    super.initState();
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

  Future<void> _startNFCRead() async {
    setState(() {
      _state = NFCVerifyState.reading;
      _errorMessage = null;
    });

    await _nfcService.startReadSession(
      onResult: (success, data) async {
        if (success) {
          await _verifyVP(data);
        } else {
          setState(() {
            _state = NFCVerifyState.error;
            _errorMessage = data;
          });
        }
      },
    );
  }

  Future<void> _verifyVP(String vpData) async {
    setState(() => _state = NFCVerifyState.verifying);

    try {
      final sdkService = ref.read(sdkServiceProvider);

      // Verify the VP using SDK
      final result = await sdkService.verifyVPNFC(vpData);

      setState(() {
        _state = NFCVerifyState.success;
        _verificationResult = result;
      });
    } catch (e) {
      setState(() {
        _state = NFCVerifyState.error;
        _errorMessage = '驗證失敗: $e';
      });
    }
  }

  void _reset() {
    _nfcService.stopSession();
    setState(() {
      _state = NFCVerifyState.idle;
      _errorMessage = null;
      _verificationResult = null;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('NFC 驗證憑證'),
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    switch (_state) {
      case NFCVerifyState.idle:
        return _buildIdleState();
      case NFCVerifyState.reading:
        return _buildReadingState();
      case NFCVerifyState.verifying:
        return _buildVerifyingState();
      case NFCVerifyState.success:
        return _buildSuccessState();
      case NFCVerifyState.error:
        return _buildErrorState();
    }
  }

  Widget _buildIdleState() {
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              width: 160,
              height: 160,
              decoration: BoxDecoration(
                color: _isNFCAvailable
                    ? AppColors.secondaryLight.withOpacity(0.2)
                    : AppColors.errorLight,
                shape: BoxShape.circle,
              ),
              child: Icon(
                Icons.nfc,
                size: 80,
                color: _isNFCAvailable ? AppColors.secondary : AppColors.error,
              ),
            ),
            const SizedBox(height: 40),
            Text(
              _isNFCAvailable ? 'NFC 驗證' : 'NFC 不可用',
              style: const TextStyle(
                fontSize: 28,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 16),
            Text(
              _isNFCAvailable
                  ? '讀取對方的 NFC 標籤來驗證其憑證'
                  : '此裝置不支援 NFC 功能',
              style: const TextStyle(
                color: AppColors.textSecondary,
                fontSize: 16,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 48),
            if (_isNFCAvailable) ...[
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: AppColors.surfaceVariant,
                  borderRadius: BorderRadius.circular(16),
                ),
                child: Column(
                  children: [
                    _buildStep('1', '點擊開始驗證'),
                    const SizedBox(height: 16),
                    _buildStep('2', '將裝置靠近對方的 NFC 標籤'),
                    const SizedBox(height: 16),
                    _buildStep('3', '等待驗證結果'),
                  ],
                ),
              ),
            ],
            const Spacer(),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton.icon(
                onPressed: _isNFCAvailable ? _startNFCRead : null,
                icon: const Icon(Icons.nfc),
                label: const Text('開始驗證'),
                style: ElevatedButton.styleFrom(
                  backgroundColor: AppColors.secondary,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildStep(String number, String text) {
    return Row(
      children: [
        Container(
          width: 32,
          height: 32,
          decoration: const BoxDecoration(
            color: AppColors.secondary,
            shape: BoxShape.circle,
          ),
          child: Center(
            child: Text(
              number,
              style: const TextStyle(
                color: Colors.white,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
        ),
        const SizedBox(width: 16),
        Text(
          text,
          style: const TextStyle(fontSize: 16),
        ),
      ],
    );
  }

  Widget _buildReadingState() {
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
                      color: AppColors.secondary.withOpacity(0.1),
                      shape: BoxShape.circle,
                    ),
                    child: Center(
                      child: Container(
                        width: 120,
                        height: 120,
                        decoration: BoxDecoration(
                          color: AppColors.secondary.withOpacity(0.2),
                          shape: BoxShape.circle,
                        ),
                        child: const Icon(
                          Icons.nfc,
                          size: 64,
                          color: AppColors.secondary,
                        ),
                      ),
                    ),
                  ),
                );
              },
            ),
            const SizedBox(height: 48),
            const Text(
              '請將裝置靠近 NFC 標籤',
              style: TextStyle(
                fontSize: 24,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 16),
            const Text(
              '正在等待讀取憑證資料...',
              style: TextStyle(
                color: AppColors.textSecondary,
                fontSize: 16,
              ),
            ),
            const SizedBox(height: 48),
            const CircularProgressIndicator(
              color: AppColors.secondary,
            ),
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

  Widget _buildVerifyingState() {
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              width: 120,
              height: 120,
              decoration: BoxDecoration(
                color: AppColors.infoLight,
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.verified_user,
                size: 64,
                color: AppColors.info,
              ),
            ),
            const SizedBox(height: 32),
            const Text(
              '正在驗證憑證...',
              style: TextStyle(
                fontSize: 24,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 24),
            const CircularProgressIndicator(
              color: AppColors.info,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSuccessState() {
    return SafeArea(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          children: [
            Container(
              width: 100,
              height: 100,
              decoration: const BoxDecoration(
                color: AppColors.successLight,
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.verified,
                size: 56,
                color: AppColors.success,
              ),
            ),
            const SizedBox(height: 24),
            const Text(
              '驗證成功！',
              style: TextStyle(
                fontSize: 28,
                fontWeight: FontWeight.bold,
                color: AppColors.success,
              ),
            ),
            const SizedBox(height: 8),
            const Text(
              '此憑證有效且可信任',
              style: TextStyle(
                color: AppColors.textSecondary,
                fontSize: 16,
              ),
            ),
            const SizedBox(height: 32),
            if (_verificationResult != null) _buildVerificationDetails(),
            const SizedBox(height: 32),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: () => context.go(AppRoutes.home),
                child: const Text('完成'),
              ),
            ),
            const SizedBox(height: 12),
            SizedBox(
              width: double.infinity,
              child: OutlinedButton(
                onPressed: _reset,
                child: const Text('驗證另一個憑證'),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildVerificationDetails() {
    return Card(
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
      ),
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Row(
              children: [
                Icon(Icons.info_outline, color: AppColors.primary),
                SizedBox(width: 8),
                Text(
                  '憑證資訊',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
            const Divider(height: 24),
            if (_verificationResult != null) ...[
              _buildDetailRow(
                '憑證類型',
                _verificationResult!['credentialType']?.toString() ?? '未知',
              ),
              _buildDetailRow(
                '發行者',
                _verificationResult!['issuer']?.toString() ?? '未知',
              ),
              _buildDetailRow(
                '持有者',
                _verificationResult!['holder']?.toString() ?? '未知',
              ),
              _buildDetailRow(
                '驗證時間',
                DateTime.now().toString().substring(0, 19),
              ),
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: AppColors.successLight,
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Row(
                  children: [
                    Icon(Icons.check_circle, color: AppColors.success, size: 20),
                    SizedBox(width: 8),
                    Text(
                      '簽章驗證通過',
                      style: TextStyle(
                        color: AppColors.success,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildDetailRow(String label, String value) {
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
              '驗證失敗',
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
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: _reset,
                child: const Text('重試'),
              ),
            ),
            const SizedBox(height: 12),
            TextButton(
              onPressed: () => context.pop(),
              child: const Text('返回'),
            ),
          ],
        ),
      ),
    );
  }
}

enum NFCVerifyState {
  idle,
  reading,
  verifying,
  success,
  error,
}
