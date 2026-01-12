// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'activity_log.dart';

// **************************************************************************
// TypeAdapterGenerator
// **************************************************************************

class ActivityTypeAdapter extends TypeAdapter<ActivityType> {
  @override
  final int typeId = 4;

  @override
  ActivityType read(BinaryReader reader) {
    switch (reader.readByte()) {
      case 0:
        return ActivityType.credentialReceived;
      case 1:
        return ActivityType.credentialPresented;
      case 2:
        return ActivityType.credentialVerified;
      case 3:
        return ActivityType.credentialTransferred;
      case 4:
        return ActivityType.credentialDeleted;
      case 5:
        return ActivityType.identityCreated;
      case 6:
        return ActivityType.presentationRejected;
      default:
        return ActivityType.credentialReceived;
    }
  }

  @override
  void write(BinaryWriter writer, ActivityType obj) {
    switch (obj) {
      case ActivityType.credentialReceived:
        writer.writeByte(0);
        break;
      case ActivityType.credentialPresented:
        writer.writeByte(1);
        break;
      case ActivityType.credentialVerified:
        writer.writeByte(2);
        break;
      case ActivityType.credentialTransferred:
        writer.writeByte(3);
        break;
      case ActivityType.credentialDeleted:
        writer.writeByte(4);
        break;
      case ActivityType.identityCreated:
        writer.writeByte(5);
        break;
      case ActivityType.presentationRejected:
        writer.writeByte(6);
        break;
    }
  }

  @override
  int get hashCode => typeId.hashCode;

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is ActivityTypeAdapter &&
          runtimeType == other.runtimeType &&
          typeId == other.typeId;
}

class ActivityLogAdapter extends TypeAdapter<ActivityLog> {
  @override
  final int typeId = 5;

  @override
  ActivityLog read(BinaryReader reader) {
    final numOfFields = reader.readByte();
    final fields = <int, dynamic>{
      for (int i = 0; i < numOfFields; i++) reader.readByte(): reader.read(),
    };
    return ActivityLog(
      id: fields[0] as String,
      type: fields[1] as ActivityType,
      timestamp: fields[2] as DateTime,
      credentialId: fields[3] as String?,
      credentialType: fields[4] as String?,
      counterparty: fields[5] as String?,
      description: fields[6] as String,
      success: fields[7] as bool,
      errorMessage: fields[8] as String?,
    );
  }

  @override
  void write(BinaryWriter writer, ActivityLog obj) {
    writer
      ..writeByte(9)
      ..writeByte(0)
      ..write(obj.id)
      ..writeByte(1)
      ..write(obj.type)
      ..writeByte(2)
      ..write(obj.timestamp)
      ..writeByte(3)
      ..write(obj.credentialId)
      ..writeByte(4)
      ..write(obj.credentialType)
      ..writeByte(5)
      ..write(obj.counterparty)
      ..writeByte(6)
      ..write(obj.description)
      ..writeByte(7)
      ..write(obj.success)
      ..writeByte(8)
      ..write(obj.errorMessage);
  }

  @override
  int get hashCode => typeId.hashCode;

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is ActivityLogAdapter &&
          runtimeType == other.runtimeType &&
          typeId == other.typeId;
}
