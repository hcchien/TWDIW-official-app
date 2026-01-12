// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'credential.dart';

// **************************************************************************
// TypeAdapterGenerator
// **************************************************************************

class CredentialStatusAdapter extends TypeAdapter<CredentialStatus> {
  @override
  final int typeId = 0;

  @override
  CredentialStatus read(BinaryReader reader) {
    switch (reader.readByte()) {
      case 0:
        return CredentialStatus.active;
      case 1:
        return CredentialStatus.suspended;
      case 2:
        return CredentialStatus.revoked;
      case 3:
        return CredentialStatus.expired;
      case 4:
        return CredentialStatus.unknown;
      default:
        return CredentialStatus.unknown;
    }
  }

  @override
  void write(BinaryWriter writer, CredentialStatus obj) {
    switch (obj) {
      case CredentialStatus.active:
        writer.writeByte(0);
        break;
      case CredentialStatus.suspended:
        writer.writeByte(1);
        break;
      case CredentialStatus.revoked:
        writer.writeByte(2);
        break;
      case CredentialStatus.expired:
        writer.writeByte(3);
        break;
      case CredentialStatus.unknown:
        writer.writeByte(4);
        break;
    }
  }

  @override
  int get hashCode => typeId.hashCode;

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is CredentialStatusAdapter &&
          runtimeType == other.runtimeType &&
          typeId == other.typeId;
}

class CredentialFieldAdapter extends TypeAdapter<CredentialField> {
  @override
  final int typeId = 1;

  @override
  CredentialField read(BinaryReader reader) {
    final numOfFields = reader.readByte();
    final fields = <int, dynamic>{
      for (int i = 0; i < numOfFields; i++) reader.readByte(): reader.read(),
    };
    return CredentialField(
      key: fields[0] as String,
      label: fields[1] as String,
      value: fields[2] as String,
      isDisclosable: fields[3] as bool,
    );
  }

  @override
  void write(BinaryWriter writer, CredentialField obj) {
    writer
      ..writeByte(4)
      ..writeByte(0)
      ..write(obj.key)
      ..writeByte(1)
      ..write(obj.label)
      ..writeByte(2)
      ..write(obj.value)
      ..writeByte(3)
      ..write(obj.isDisclosable);
  }

  @override
  int get hashCode => typeId.hashCode;

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is CredentialFieldAdapter &&
          runtimeType == other.runtimeType &&
          typeId == other.typeId;
}

class CredentialDisplayAdapter extends TypeAdapter<CredentialDisplay> {
  @override
  final int typeId = 2;

  @override
  CredentialDisplay read(BinaryReader reader) {
    final numOfFields = reader.readByte();
    final fields = <int, dynamic>{
      for (int i = 0; i < numOfFields; i++) reader.readByte(): reader.read(),
    };
    return CredentialDisplay(
      name: fields[0] as String?,
      description: fields[1] as String?,
      backgroundColor: fields[2] as String?,
      textColor: fields[3] as String?,
      logoUrl: fields[4] as String?,
      backgroundImageUrl: fields[5] as String?,
    );
  }

  @override
  void write(BinaryWriter writer, CredentialDisplay obj) {
    writer
      ..writeByte(6)
      ..writeByte(0)
      ..write(obj.name)
      ..writeByte(1)
      ..write(obj.description)
      ..writeByte(2)
      ..write(obj.backgroundColor)
      ..writeByte(3)
      ..write(obj.textColor)
      ..writeByte(4)
      ..write(obj.logoUrl)
      ..writeByte(5)
      ..write(obj.backgroundImageUrl);
  }

  @override
  int get hashCode => typeId.hashCode;

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is CredentialDisplayAdapter &&
          runtimeType == other.runtimeType &&
          typeId == other.typeId;
}

class StoredCredentialAdapter extends TypeAdapter<StoredCredential> {
  @override
  final int typeId = 3;

  @override
  StoredCredential read(BinaryReader reader) {
    final numOfFields = reader.readByte();
    final fields = <int, dynamic>{
      for (int i = 0; i < numOfFields; i++) reader.readByte(): reader.read(),
    };
    return StoredCredential(
      id: fields[0] as String,
      rawToken: fields[1] as String,
      credentialType: fields[2] as String,
      issuer: fields[3] as String,
      issuerName: fields[4] as String,
      issuedAt: fields[5] as DateTime,
      expiresAt: fields[6] as DateTime,
      fields: (fields[7] as List).cast<CredentialField>(),
      display: fields[8] as CredentialDisplay?,
      status: fields[9] as CredentialStatus,
      addedAt: fields[10] as DateTime,
      subjectDid: fields[11] as String?,
    );
  }

  @override
  void write(BinaryWriter writer, StoredCredential obj) {
    writer
      ..writeByte(12)
      ..writeByte(0)
      ..write(obj.id)
      ..writeByte(1)
      ..write(obj.rawToken)
      ..writeByte(2)
      ..write(obj.credentialType)
      ..writeByte(3)
      ..write(obj.issuer)
      ..writeByte(4)
      ..write(obj.issuerName)
      ..writeByte(5)
      ..write(obj.issuedAt)
      ..writeByte(6)
      ..write(obj.expiresAt)
      ..writeByte(7)
      ..write(obj.fields)
      ..writeByte(8)
      ..write(obj.display)
      ..writeByte(9)
      ..write(obj.status)
      ..writeByte(10)
      ..write(obj.addedAt)
      ..writeByte(11)
      ..write(obj.subjectDid);
  }

  @override
  int get hashCode => typeId.hashCode;

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is StoredCredentialAdapter &&
          runtimeType == other.runtimeType &&
          typeId == other.typeId;
}
