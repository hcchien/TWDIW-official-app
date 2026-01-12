import 'package:equatable/equatable.dart';
import 'package:hive/hive.dart';

part 'credential.g.dart';

/// Credential status enum
@HiveType(typeId: 0)
enum CredentialStatus {
  @HiveField(0)
  active,
  @HiveField(1)
  suspended,
  @HiveField(2)
  revoked,
  @HiveField(3)
  expired,
  @HiveField(4)
  unknown,
}

/// A field within a credential that can be selectively disclosed
@HiveType(typeId: 1)
class CredentialField extends Equatable {
  @HiveField(0)
  final String key;

  @HiveField(1)
  final String label;

  @HiveField(2)
  final String value;

  @HiveField(3)
  final bool isDisclosable;

  const CredentialField({
    required this.key,
    required this.label,
    required this.value,
    this.isDisclosable = true,
  });

  factory CredentialField.fromJson(Map<String, dynamic> json) {
    return CredentialField(
      key: json['key'] as String? ?? '',
      label: json['label'] as String? ?? json['key'] as String? ?? '',
      value: json['value']?.toString() ?? '',
      isDisclosable: json['isDisclosable'] as bool? ?? true,
    );
  }

  Map<String, dynamic> toJson() => {
        'key': key,
        'label': label,
        'value': value,
        'isDisclosable': isDisclosable,
      };

  @override
  List<Object?> get props => [key, label, value, isDisclosable];
}

/// Credential display metadata (colors, images, etc.)
@HiveType(typeId: 2)
class CredentialDisplay extends Equatable {
  @HiveField(0)
  final String? name;

  @HiveField(1)
  final String? description;

  @HiveField(2)
  final String? backgroundColor;

  @HiveField(3)
  final String? textColor;

  @HiveField(4)
  final String? logoUrl;

  @HiveField(5)
  final String? backgroundImageUrl;

  const CredentialDisplay({
    this.name,
    this.description,
    this.backgroundColor,
    this.textColor,
    this.logoUrl,
    this.backgroundImageUrl,
  });

  factory CredentialDisplay.fromJson(Map<String, dynamic> json) {
    return CredentialDisplay(
      name: json['name'] as String?,
      description: json['description'] as String?,
      backgroundColor: json['background_color'] as String?,
      textColor: json['text_color'] as String?,
      logoUrl: json['logo']?['url'] as String?,
      backgroundImageUrl: json['background_image']?['url'] as String?,
    );
  }

  Map<String, dynamic> toJson() => {
        'name': name,
        'description': description,
        'background_color': backgroundColor,
        'text_color': textColor,
        'logo': logoUrl != null ? {'url': logoUrl} : null,
        'background_image':
            backgroundImageUrl != null ? {'url': backgroundImageUrl} : null,
      };

  @override
  List<Object?> get props => [
        name,
        description,
        backgroundColor,
        textColor,
        logoUrl,
        backgroundImageUrl,
      ];
}

/// Stored Verifiable Credential
@HiveType(typeId: 3)
class StoredCredential extends Equatable {
  @HiveField(0)
  final String id;

  @HiveField(1)
  final String rawToken;

  @HiveField(2)
  final String credentialType;

  @HiveField(3)
  final String issuer;

  @HiveField(4)
  final String issuerName;

  @HiveField(5)
  final DateTime issuedAt;

  @HiveField(6)
  final DateTime expiresAt;

  @HiveField(7)
  final List<CredentialField> fields;

  @HiveField(8)
  final CredentialDisplay? display;

  @HiveField(9)
  final CredentialStatus status;

  @HiveField(10)
  final DateTime addedAt;

  @HiveField(11)
  final String? subjectDid;

  const StoredCredential({
    required this.id,
    required this.rawToken,
    required this.credentialType,
    required this.issuer,
    required this.issuerName,
    required this.issuedAt,
    required this.expiresAt,
    required this.fields,
    this.display,
    this.status = CredentialStatus.active,
    required this.addedAt,
    this.subjectDid,
  });

  bool get isExpired => DateTime.now().isAfter(expiresAt);

  bool get isValid =>
      status == CredentialStatus.active && !isExpired;

  StoredCredential copyWith({
    String? id,
    String? rawToken,
    String? credentialType,
    String? issuer,
    String? issuerName,
    DateTime? issuedAt,
    DateTime? expiresAt,
    List<CredentialField>? fields,
    CredentialDisplay? display,
    CredentialStatus? status,
    DateTime? addedAt,
    String? subjectDid,
  }) {
    return StoredCredential(
      id: id ?? this.id,
      rawToken: rawToken ?? this.rawToken,
      credentialType: credentialType ?? this.credentialType,
      issuer: issuer ?? this.issuer,
      issuerName: issuerName ?? this.issuerName,
      issuedAt: issuedAt ?? this.issuedAt,
      expiresAt: expiresAt ?? this.expiresAt,
      fields: fields ?? this.fields,
      display: display ?? this.display,
      status: status ?? this.status,
      addedAt: addedAt ?? this.addedAt,
      subjectDid: subjectDid ?? this.subjectDid,
    );
  }

  factory StoredCredential.fromJson(Map<String, dynamic> json) {
    return StoredCredential(
      id: json['id'] as String,
      rawToken: json['rawToken'] as String,
      credentialType: json['credentialType'] as String,
      issuer: json['issuer'] as String,
      issuerName: json['issuerName'] as String? ?? 'Unknown Issuer',
      issuedAt: DateTime.parse(json['issuedAt'] as String),
      expiresAt: DateTime.parse(json['expiresAt'] as String),
      fields: (json['fields'] as List<dynamic>?)
              ?.map((e) => CredentialField.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      display: json['display'] != null
          ? CredentialDisplay.fromJson(json['display'] as Map<String, dynamic>)
          : null,
      status: CredentialStatus.values.firstWhere(
        (e) => e.name == json['status'],
        orElse: () => CredentialStatus.unknown,
      ),
      addedAt: DateTime.parse(json['addedAt'] as String),
      subjectDid: json['subjectDid'] as String?,
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'rawToken': rawToken,
        'credentialType': credentialType,
        'issuer': issuer,
        'issuerName': issuerName,
        'issuedAt': issuedAt.toIso8601String(),
        'expiresAt': expiresAt.toIso8601String(),
        'fields': fields.map((e) => e.toJson()).toList(),
        'display': display?.toJson(),
        'status': status.name,
        'addedAt': addedAt.toIso8601String(),
        'subjectDid': subjectDid,
      };

  @override
  List<Object?> get props => [
        id,
        rawToken,
        credentialType,
        issuer,
        issuerName,
        issuedAt,
        expiresAt,
        fields,
        display,
        status,
        addedAt,
        subjectDid,
      ];
}
