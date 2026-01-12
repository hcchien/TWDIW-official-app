import 'package:equatable/equatable.dart';

/// A credential requested by a verifier
class RequestedCredential extends Equatable {
  final String cardType;
  final String? cardId;
  final String? name;
  final List<String> requiredFields;
  final List<String> optionalFields;
  final String? group;
  final String? rule;
  final int? count;
  final int? max;

  const RequestedCredential({
    required this.cardType,
    this.cardId,
    this.name,
    required this.requiredFields,
    this.optionalFields = const [],
    this.group,
    this.rule,
    this.count,
    this.max,
  });

  factory RequestedCredential.fromJson(Map<String, dynamic> json) {
    return RequestedCredential(
      cardType: json['card_type'] as String? ?? '',
      cardId: json['card_id'] as String?,
      name: json['name'] as String?,
      requiredFields: (json['required_fields'] as List<dynamic>?)
              ?.map((e) => e.toString())
              .toList() ??
          [],
      optionalFields: (json['optional_fields'] as List<dynamic>?)
              ?.map((e) => e.toString())
              .toList() ??
          [],
      group: json['group'] as String?,
      rule: json['rule'] as String?,
      count: json['count'] as int?,
      max: json['max'] as int?,
    );
  }

  Map<String, dynamic> toJson() => {
        'card_type': cardType,
        'card_id': cardId,
        'name': name,
        'required_fields': requiredFields,
        'optional_fields': optionalFields,
        'group': group,
        'rule': rule,
        'count': count,
        'max': max,
      };

  @override
  List<Object?> get props => [
        cardType,
        cardId,
        name,
        requiredFields,
        optionalFields,
        group,
        rule,
        count,
        max,
      ];
}

/// Parsed VP request from a verifier's QR code
class PresentationRequest extends Equatable {
  final String requestToken;
  final List<RequestedCredential> requestedCredentials;
  final String? verifierName;
  final String? verifierDid;
  final String? purpose;
  final String? nonce;
  final String? responseUri;

  const PresentationRequest({
    required this.requestToken,
    required this.requestedCredentials,
    this.verifierName,
    this.verifierDid,
    this.purpose,
    this.nonce,
    this.responseUri,
  });

  factory PresentationRequest.fromJson(Map<String, dynamic> json) {
    final List<RequestedCredential> credentials = [];

    // Parse presentation definition
    final presentationDef = json['presentation_definition'];
    if (presentationDef != null) {
      final inputDescriptors =
          presentationDef['input_descriptors'] as List<dynamic>?;
      if (inputDescriptors != null) {
        for (final descriptor in inputDescriptors) {
          final constraints = descriptor['constraints'] as Map<String, dynamic>?;
          final fields = constraints?['fields'] as List<dynamic>?;

          final requiredFields = <String>[];
          final optionalFields = <String>[];

          if (fields != null) {
            for (final field in fields) {
              final path = (field['path'] as List<dynamic>?)?.firstOrNull;
              if (path != null) {
                // Extract field name from path like "$.credentialSubject.name"
                final parts = path.toString().split('.');
                final fieldName = parts.length > 2 ? parts.last : path.toString();

                if (field['optional'] == true) {
                  optionalFields.add(fieldName);
                } else {
                  requiredFields.add(fieldName);
                }
              }
            }
          }

          credentials.add(RequestedCredential(
            cardType: descriptor['id']?.toString() ?? '',
            name: descriptor['name']?.toString(),
            requiredFields: requiredFields,
            optionalFields: optionalFields,
            group: descriptor['group']?.toString(),
          ));
        }
      }
    }

    // Also check for direct credential requests
    final directRequests = json['requested_credentials'] as List<dynamic>?;
    if (directRequests != null) {
      for (final request in directRequests) {
        credentials.add(
            RequestedCredential.fromJson(request as Map<String, dynamic>));
      }
    }

    return PresentationRequest(
      requestToken: json['request_token'] as String? ?? '',
      requestedCredentials: credentials,
      verifierName: json['verifier_name'] as String? ??
          json['client_id'] as String?,
      verifierDid: json['verifier_did'] as String?,
      purpose: json['purpose'] as String?,
      nonce: json['nonce'] as String?,
      responseUri: json['response_uri'] as String?,
    );
  }

  Map<String, dynamic> toJson() => {
        'request_token': requestToken,
        'requested_credentials':
            requestedCredentials.map((e) => e.toJson()).toList(),
        'verifier_name': verifierName,
        'verifier_did': verifierDid,
        'purpose': purpose,
        'nonce': nonce,
        'response_uri': responseUri,
      };

  @override
  List<Object?> get props => [
        requestToken,
        requestedCredentials,
        verifierName,
        verifierDid,
        purpose,
        nonce,
        responseUri,
      ];
}
