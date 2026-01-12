import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../../../data/models/credential.dart';
import '../../../core/theme/app_colors.dart';
import '../../widgets/credential_card.dart';
import '../../router/app_router.dart';

class ReceiveSuccessScreen extends StatefulWidget {
  final StoredCredential? credential;

  const ReceiveSuccessScreen({super.key, this.credential});

  @override
  State<ReceiveSuccessScreen> createState() => _ReceiveSuccessScreenState();
}

class _ReceiveSuccessScreenState extends State<ReceiveSuccessScreen>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _scaleAnimation;
  late Animation<double> _fadeAnimation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      duration: const Duration(milliseconds: 800),
      vsync: this,
    );

    _scaleAnimation = Tween<double>(begin: 0.0, end: 1.0).animate(
      CurvedAnimation(
        parent: _controller,
        curve: const Interval(0.0, 0.5, curve: Curves.elasticOut),
      ),
    );

    _fadeAnimation = Tween<double>(begin: 0.0, end: 1.0).animate(
      CurvedAnimation(
        parent: _controller,
        curve: const Interval(0.3, 1.0, curve: Curves.easeOut),
      ),
    );

    _controller.forward();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            children: [
              const Spacer(),
              AnimatedBuilder(
                animation: _controller,
                builder: (context, child) {
                  return Transform.scale(
                    scale: _scaleAnimation.value,
                    child: Container(
                      width: 100,
                      height: 100,
                      decoration: const BoxDecoration(
                        color: AppColors.successLight,
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(
                        Icons.check,
                        size: 56,
                        color: AppColors.success,
                      ),
                    ),
                  );
                },
              ),
              const SizedBox(height: 32),
              FadeTransition(
                opacity: _fadeAnimation,
                child: Column(
                  children: [
                    const Text(
                      '憑證取得成功！',
                      style: TextStyle(
                        fontSize: 28,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 12),
                    const Text(
                      '您的數位憑證已安全儲存至皮夾',
                      style: TextStyle(
                        color: AppColors.textSecondary,
                        fontSize: 14,
                      ),
                    ),
                    if (widget.credential != null) ...[
                      const SizedBox(height: 32),
                      AspectRatio(
                        aspectRatio: 1.6,
                        child: CredentialCard(
                          credential: widget.credential!,
                        ),
                      ),
                    ],
                  ],
                ),
              ),
              const Spacer(),
              FadeTransition(
                opacity: _fadeAnimation,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    ElevatedButton(
                      onPressed: () => context.go(AppRoutes.home),
                      child: const Text('返回皮夾'),
                    ),
                    const SizedBox(height: 12),
                    OutlinedButton(
                      onPressed: () {
                        if (widget.credential != null) {
                          context.push(
                            '/credential/${widget.credential!.id}',
                            extra: widget.credential,
                          );
                        }
                      },
                      child: const Text('查看憑證詳情'),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
