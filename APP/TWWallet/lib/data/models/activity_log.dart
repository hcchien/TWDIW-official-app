import 'package:equatable/equatable.dart';
import 'package:hive/hive.dart';

part 'activity_log.g.dart';

/// Type of activity performed
@HiveType(typeId: 4)
enum ActivityType {
  @HiveField(0)
  credentialReceived,
  @HiveField(1)
  credentialPresented,
  @HiveField(2)
  credentialVerified,
  @HiveField(3)
  credentialTransferred,
  @HiveField(4)
  credentialDeleted,
  @HiveField(5)
  identityCreated,
  @HiveField(6)
  presentationRejected,
}

/// Activity log entry
@HiveType(typeId: 5)
class ActivityLog extends Equatable {
  @HiveField(0)
  final String id;

  @HiveField(1)
  final ActivityType type;

  @HiveField(2)
  final DateTime timestamp;

  @HiveField(3)
  final String? credentialId;

  @HiveField(4)
  final String? credentialType;

  @HiveField(5)
  final String? counterparty;

  @HiveField(6)
  final String description;

  @HiveField(7)
  final bool success;

  @HiveField(8)
  final String? errorMessage;

  const ActivityLog({
    required this.id,
    required this.type,
    required this.timestamp,
    this.credentialId,
    this.credentialType,
    this.counterparty,
    required this.description,
    this.success = true,
    this.errorMessage,
  });

  String get typeDisplayName {
    switch (type) {
      case ActivityType.credentialReceived:
        return '取得憑證';
      case ActivityType.credentialPresented:
        return '出示憑證';
      case ActivityType.credentialVerified:
        return '驗證憑證';
      case ActivityType.credentialTransferred:
        return '轉移憑證';
      case ActivityType.credentialDeleted:
        return '刪除憑證';
      case ActivityType.identityCreated:
        return '建立身分';
      case ActivityType.presentationRejected:
        return '拒絕出示';
    }
  }

  String get typeIcon {
    switch (type) {
      case ActivityType.credentialReceived:
        return '📥';
      case ActivityType.credentialPresented:
        return '📤';
      case ActivityType.credentialVerified:
        return '✅';
      case ActivityType.credentialTransferred:
        return '↗️';
      case ActivityType.credentialDeleted:
        return '🗑️';
      case ActivityType.identityCreated:
        return '🆔';
      case ActivityType.presentationRejected:
        return '❌';
    }
  }

  factory ActivityLog.fromJson(Map<String, dynamic> json) {
    return ActivityLog(
      id: json['id'] as String,
      type: ActivityType.values.firstWhere(
        (e) => e.name == json['type'],
        orElse: () => ActivityType.credentialReceived,
      ),
      timestamp: DateTime.parse(json['timestamp'] as String),
      credentialId: json['credentialId'] as String?,
      credentialType: json['credentialType'] as String?,
      counterparty: json['counterparty'] as String?,
      description: json['description'] as String,
      success: json['success'] as bool? ?? true,
      errorMessage: json['errorMessage'] as String?,
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'type': type.name,
        'timestamp': timestamp.toIso8601String(),
        'credentialId': credentialId,
        'credentialType': credentialType,
        'counterparty': counterparty,
        'description': description,
        'success': success,
        'errorMessage': errorMessage,
      };

  @override
  List<Object?> get props => [
        id,
        type,
        timestamp,
        credentialId,
        credentialType,
        counterparty,
        description,
        success,
        errorMessage,
      ];
}
