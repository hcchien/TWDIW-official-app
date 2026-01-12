import 'dart:convert';
import 'dart:typed_data';
import 'package:did_sdk_module/did_key_gen.dart';
import 'package:archive/archive.dart';
import 'package:pointycastle/export.dart';
import 'package:uuid/uuid.dart';
import 'utils/utils.dart';
import 'http_service.dart';

/// OpenidVcVp class
/// This class encapsulates OpenID VC/VP related data and logic.
class OpenidVcVp {
  HttpService httpService = HttpService();
  Utils utils = Utils();

  /// Sign JWT using pure Dart implementation
  Future<String> _signJwt(Map<String, dynamic> header, Map<String, dynamic> payload) async {
    // Use async version to ensure key is loaded from storage if needed
    final keyPair = await DIDKeyGenerator.getKeyPairAsync();
    if (keyPair == null) {
      throw Exception('Key not initialized. Call initKx first.');
    }

    final headerJson = jsonEncode(header);
    final payloadJson = jsonEncode(payload);

    final headerBase64 = _base64UrlEncode(utf8.encode(headerJson));
    final payloadBase64 = _base64UrlEncode(utf8.encode(payloadJson));

    final signingInput = '$headerBase64.$payloadBase64';
    final signingInputBytes = Uint8List.fromList(utf8.encode(signingInput));

    // Hash the signing input with SHA256
    final digest = SHA256Digest();
    final hash = digest.process(signingInputBytes);

    // Sign with ECDSA
    final privateKey = keyPair.privateKey as ECPrivateKey;
    final signer = ECDSASigner(null)
      ..init(true, PrivateKeyParameter<ECPrivateKey>(privateKey));

    final signature = signer.generateSignature(hash) as ECSignature;

    // Convert signature to fixed-length format (R || S, each 32 bytes)
    final r = _bigIntToBytes(signature.r, 32);
    final s = _bigIntToBytes(signature.s, 32);
    final signatureBytes = Uint8List.fromList([...r, ...s]);

    final signatureBase64 = _base64UrlEncode(signatureBytes);

    return '$headerBase64.$payloadBase64.$signatureBase64';
  }

  String _base64UrlEncode(List<int> bytes) {
    return base64Url.encode(bytes).replaceAll('=', '');
  }

  Uint8List _bigIntToBytes(BigInt bigInt, int length) {
    var hexString = bigInt.toRadixString(16);
    if (hexString.length % 2 != 0) {
      hexString = '0$hexString';
    }
    final bytes = <int>[];
    for (var i = 0; i < hexString.length; i += 2) {
      bytes.add(int.parse(hexString.substring(i, i + 2), radix: 16));
    }
    if (bytes.length >= length) {
      return Uint8List.fromList(bytes.sublist(bytes.length - length));
    }
    final padded = Uint8List(length);
    padded.setRange(length - bytes.length, length, bytes);
    return padded;
  }

//VC
  /// applyVC method
  /// This method is responsible for applying for VC
  Future<String> applyVCKx(String didFile, String qrCode, String txCode) async {
    try {
      final encode103i = qrCode.split('credential_offer_uri=')[1];
      final decoded103i = Uri.decodeFull(encode103i);

      // dwissuer-oidvci-101i
      Map<String, dynamic> credentialObject =
          await httpService.getCredentialObject(decoded103i);
      if (credentialObject['error'] != null) {
        return utils.response(code: '2011', message: credentialObject['error']);
      }
      // dwissuer-oidvci-102i
      Map<String, dynamic> credentialMetadata =
          await httpService.getCredentialMetadata(
              credentialObject['applyVCUrl'],
              credentialObject['credential_configuration_ids'][0]);
      if (credentialMetadata['error'] != null) {
        return utils.response(
            code: '2012', message: credentialMetadata['error']);
      }
      // dwissuer-oidvci-103i
      Map<String, dynamic> credentialDefinition =
          credentialMetadata['credentialMetadata']
                  ['credential_configurations_supported']
              [credentialObject['credential_configuration_ids'][0]];

      // dwissuer-oidvci-104i
      Map<String, dynamic> accessToken = await httpService.getAccessToken(
          credentialObject['applyVCUrl'],
          credentialObject['pre-authorized_code'],
          credentialObject['credential_configuration_ids'][0],
          txCode);
      if (accessToken['error'] != null) {
        return utils.response(code: '2013', message: accessToken['error']);
      }

      Map<String, dynamic> didMap = jsonDecode(didFile);

      final header = {
        'alg': 'ES256',
        'typ': 'openid4vci-proof+jwt',
        'kid': didMap['id'],
      };

      final payload = {
        'iss': 'moda_dw', // issuer
        'aud': credentialObject['applyVCUrl'], // audience
        'iat': DateTime.now().millisecondsSinceEpoch ~/
            1000, // issued at (current time in seconds)
        'nonce': accessToken['c_nonce'] // unique nonce
      };

      String jwtToken = await _signJwt(header, payload);

      // dwissuer-oidvci-105i
      Map<String, dynamic> credential = await httpService.getVC(
          credentialObject['applyVCUrl'],
          accessToken['access_token'],
          credentialObject['credential_configuration_ids'][0],
          jwtToken);
      if (credential['error'] != null) {
        return utils.response(code: '2014', message: credential['error']);
      }

      final data = {
        'credential': credential['credential'],
        'credentialDefinition': credentialDefinition,
        'credentialMetadata': credentialMetadata['credentialMetadata']
      };

      return utils.response(code: '0', message: 'SUCCESS', data: data);
    } catch (e) {
      return utils.response(code: '1', message: e.toString());
    }
  }

  /// decodeVC method
  /// This method is responsible for decoding VC.
  String decodeVC(String jwToken) {
    try {
      final vcPayload = utils.jwtDecode(jwToken);
      Map<String, dynamic> field = utils.sdJwtDecode(jwToken);
      vcPayload['vc']['credentialSubject']['field'] = field;

      return utils.response(code: '0', message: 'SUCCESS', data: vcPayload);
    } catch (e) {
      return utils.response(code: '1', message: e.toString());
    }
  }

  /// verifyVC method
  /// This method is responsible for verifying VC
  Future<String> verifyVC(
      String jwToken, String didFile, String frontUrl) async {
    try {
      final vcPayload = utils.jwtDecode(jwToken);

      var trustFlg = false;
      var vcFlg = false;
      var expFlg = false;
      var holderFlg = false;
      var issFlg = false;
      var badgeFlg = false;

      //Issuer is in the list
      Map<String, dynamic> issStatus =
          await httpService.getVCIssStatus(frontUrl, vcPayload['iss']);
      if (issStatus['error'] == null) {
        if (issStatus['data']['status'] == 1) {
          trustFlg = true;
        }
        if (issStatus['data']['org'].containsKey('x509_type')) {
          badgeFlg = true;
        }
      } else if (issStatus['error'] == 9999) {
        return utils.response(code: '2', message: 'API FAIL');
      }

      // VC is valid
      if (vcPayload['vc']['credentialStatus'] is Map) {
        Map<String, dynamic> vcListToken = await httpService.getVCList(
            vcPayload['vc']['credentialStatus']['statusListCredential']);
        if (vcListToken['error'] == null) {
          vcFlg = await checkVCValid(
              vcPayload['vc']['credentialStatus'], vcListToken);
        } else if (vcListToken['error'] == 9999) {
          return utils.response(code: '2', message: 'API FAIL');
        }
      } else {
        for (var item in vcPayload['vc']['credentialStatus']) {
          Map<String, dynamic> vcListToken =
              await httpService.getVCList(item['statusListCredential']);
          if (vcListToken['error'] == null) {
            final valid = await checkVCValid(item, vcListToken);
            if (!valid) {
              vcFlg = false;
              break;
            } else {
              vcFlg = true;
            }
          } else if (vcListToken['error'] == 9999) {
            return utils.response(code: '2', message: 'API FAIL');
          }
        }
      }

      // Verify VC sub matches DID
      Map<String, dynamic> didMap = jsonDecode(didFile);
      if (vcPayload['sub'] == didMap['id']) {
        holderFlg = true;
      }

      // Verify VC is not expired
      int exp = vcPayload['exp'];
      int currentTime = DateTime.now().millisecondsSinceEpoch ~/ 1000;
      if (currentTime < exp) {
        expFlg = true;
      }

      // VC signature verification
      List<String> parts = jwToken.split('.');
      List<String> sdParts = parts[2].split('~');
      String token = '${parts[0]}.${parts[1]}.${sdParts[0]}';
      if (issStatus['data'] != null) {
        final didPayload = utils.jwtDecode(issStatus['data']['did']);
        ECPublicKey publicKey = DIDKeyGenerator().convertJwkToECPublicKey(
            didPayload['verificationMethod'][0]['publicKeyJwk']);

        issFlg = utils.verifyJwt(token, publicKey);
      }

      var data = {};
      if (trustFlg & vcFlg & expFlg & holderFlg & issFlg) {
        data = {'trust_badge': badgeFlg};
        return utils.response(code: '0', message: 'SUCCESS', data: data);
      } else {
        data = {
          'trust': trustFlg,
          'vc': vcFlg,
          'issuer': issFlg,
          'exp': expFlg,
          'holder': holderFlg,
          'trust_badge': badgeFlg
        };
        return utils.response(code: '3', message: 'Verification failed', data: data);
      }
    } catch (e) {
      return utils.response(code: '1', message: e.toString());
    }
  }

  /// verifyVCOffline method
  /// This method is responsible for offline verification
  Future<String> verifyVCOffline(
      String jwToken, String didFile, List issList, List vcList) async {
    try {
      final vcPayload = utils.jwtDecode(jwToken);

      bool? trustFlg = false;
      bool? vcFlg = false;
      bool? issFlg = false;
      var expFlg = false;
      var holderFlg = false;
      var badgeFlg = false;

      var issuer = {};
      //Issuer is in the list
      if (issList.isEmpty) {
        trustFlg = null;
      } else {
        for (var iss in issList) {
          final issPayload = utils.jwtDecode(iss['did']);
          if (vcPayload['iss'] == issPayload['id']) {
            issuer = iss;
            if (iss['status'] == 1) {
              trustFlg = true;
            }
            if (iss['org'].containsKey('x509_type')) {
              badgeFlg = true;
            }
          }
        }
      }

      // VC is valid, convert to sha256 for comparison
      final vcTokenHash = utils.sha256Hash(jwToken);
      if (vcList.isEmpty) {
        vcFlg = null;
      } else {
        for (var vcListToken in vcList) {
          if (vcListToken.containsKey(vcTokenHash)) {
            for (int i = 0; i < vcListToken[vcTokenHash].length; i++) {
              final valid = await checkVCValid(
                  vcPayload['vc']['credentialStatus'][i],
                  vcListToken[vcTokenHash][i]);
              if (!valid) {
                vcFlg = false;
                break;
              } else {
                vcFlg = true;
              }
            }
          }
        }
      }

      // Verify VC sub matches DID
      Map<String, dynamic> didMap = jsonDecode(didFile);
      if (vcPayload['sub'] == didMap['id']) {
        holderFlg = true;
      }

      // Verify VC is not expired
      int exp = vcPayload['exp'];
      int currentTime = DateTime.now().millisecondsSinceEpoch ~/ 1000;
      if (currentTime < exp) {
        expFlg = true;
      }

      // VC signature verification
      List<String> parts = jwToken.split('.');
      List<String> sdParts = parts[2].split('~');
      String token = '${parts[0]}.${parts[1]}.${sdParts[0]}';
      if (issList.isEmpty) {
        issFlg = null;
      } else {
        if (issuer['did'] != null) {
          final didPayload = utils.jwtDecode(issuer['did']);
          ECPublicKey publicKey = DIDKeyGenerator().convertJwkToECPublicKey(
              didPayload['verificationMethod'][0]['publicKeyJwk']);

          issFlg = utils.verifyJwt(token, publicKey);
        }
      }

      var data = {};
      if (trustFlg == true &&
          vcFlg == true &&
          issFlg == true &&
          expFlg &&
          holderFlg) {
        data = {'trust_badge': badgeFlg};
        return utils.response(code: '0', message: 'SUCCESS', data: data);
      } else {
        data = {
          'trust': trustFlg,
          'vc': vcFlg,
          'issuer': issFlg,
          'exp': expFlg,
          'holder': holderFlg,
          'trust_badge': badgeFlg
        };
        return utils.response(code: '3', message: 'Verification failed', data: data);
      }
    } catch (e) {
      return utils.response(code: '1', message: e.toString());
    }
  }

  /// downloadIssList method
  /// This method downloads the trust list for offline VC verification
  Future<String> downloadIssList(String url) async {
    try {
      Map<String, dynamic> issList = await httpService.getVCIssList(url, 20, 0);

      if (issList['error'] != null) {
        return utils.response(code: '2', message: issList['error']);
      }
      if (issList['count'] > 20) {
        for (int i = 0; i < issList['count'] ~/ 20; i++) {
          Map<String, dynamic> moreList =
              await httpService.getVCIssList(url, 20, i + 1);
          issList['dids'] = issList['dids'] + moreList['dids'];
        }
      }
      return utils.response(
          code: '0', message: 'SUCCESS', data: issList['dids']);
    } catch (e) {
      return utils.response(code: '1', message: e.toString());
    }
  }

  /// downloadAllVCList method
  /// This method downloads VC status list for offline VC verification
  Future<String> downloadAllVCList(List vcs, List vcList) async {
    try {
      List resultList = [];
      final vcMap = _vcListMap(vcList);

      for (var vcToken in vcs) {
        Map response = {};
        List allList = [];

        final vcPayload = utils.jwtDecode(vcToken);
        final vcTokenHash = utils.sha256Hash(vcToken);

        if (vcPayload['vc']['credentialStatus'] is Map) {
          Map<String, dynamic> vcListToken = await httpService.getVCList(
            vcPayload['vc']['credentialStatus']['statusListCredential'],
          );

          if (vcMap.containsKey(vcTokenHash) && vcListToken['error'] != null) {
            allList.add(vcMap[vcTokenHash]![0]);
          } else {
            allList.add(vcListToken);
          }
        } else {
          int index = 0;
          for (var list in vcPayload['vc']['credentialStatus']) {
            Map<String, dynamic> vcListToken = await httpService.getVCList(
              list['statusListCredential'],
            );

            if (vcMap.containsKey(vcTokenHash) &&
                vcListToken['error'] != null &&
                index < vcMap[vcTokenHash]!.length) {
              allList.add(vcMap[vcTokenHash]![index]);
            } else {
              allList.add(vcListToken);
            }
            index++;
          }
        }

        response[vcTokenHash] = allList;
        resultList.add(response);
      }
      return utils.response(code: '0', message: 'SUCCESS', data: resultList);
    } catch (e) {
      return utils.response(code: '1', message: e.toString());
    }
  }

//VP
  /// parseVPQrcode method
  /// This method parses VP request QR code
  Future<String> parseVPQrcode(String qrCode, String frontUrl) async {
    try {
      final verifier102i = qrCode.split('request_uri=')[1].split('&')[0];
      final decoded102i = Uri.decodeFull(verifier102i);

      Map<String, dynamic> objectToken =
          await httpService.getVPObject(decoded102i);
      if (objectToken['error'] != null) {
        return utils.response(code: '4011', message: objectToken['error']);
      }
      final objectPayload = utils.jwtDecode(objectToken['token']);

      var data = {};

      final requestData = parsePresentationDefinition(objectPayload);

      //Issuer is in the list
      if (objectPayload['client_id'].contains('did:')) {
        Map<String, dynamic> issStatus = await httpService.getVCIssStatus(
            frontUrl, objectPayload['client_id']);
        if (issStatus['error'] != null) {
          data = {
            'request_token': objectToken['token'],
            'request_data': requestData
          };
          return utils.response(
              code: '4012', message: 'Verifier DID Fail', data: data);
        }
      }

      data = {
        'request_token': objectToken['token'],
        'request_data': requestData
      };

      return utils.response(code: '0', message: 'SUCCESS', data: data);
    } catch (e) {
      return utils.response(code: '1', message: e.toString());
    }
  }

  /// generateVPKx method
  /// This method generates VP
  Future<String> generateVPKx(String didFile, String requestToken,
      List<dynamic> vcs, String customData) async {
    try {
      Map<String, dynamic> didMap = jsonDecode(didFile);
      final objectPayload = utils.jwtDecode(requestToken);

      List vcsResult = [];
      List descriptorMap = [];

      //Generate sdJWT and descriptor_map based on field
      vcs.asMap().forEach((index, vc) {
        Map<String, dynamic> result = {};
        result = utils.sdJwtEncode(vc['vc'], vc['field']);
        vcsResult.add(result['data']);
        //Build descriptor_map
        final cardId = (vc['card_id'] ?? '').toString();
        final id = cardId.isNotEmpty
            ? cardId
            : objectPayload['presentation_definition']['input_descriptors']
                [index]['id'];
        descriptorMap.add({
          'id': id,
          'format': 'jwt_vp',
          'path': '\$',
          'path_nested': {
            'id': id,
            'format': 'jwt_vc',
            'path': '\$.vp.verifiableCredential[$index]'
          }
        });
      });

      final header = {
        'typ': 'JWT',
        'alg': 'ES256',
        "jwk": didMap['verificationMethod'][0]['publicKeyJwk']
      };

      const uuid = Uuid();
      final payload = {
        'sub': didMap['id'], // subject
        'aud': objectPayload['client_id'], // audience
        'iss': didMap['id'], // issuer
        'nbf': DateTime.now().millisecondsSinceEpoch ~/
            1000, // issued at (current time in seconds)
        'vp': {
          'context': ['https://www.w3.org/2018/credentials/v1'],
          'type': ['VerifiablePresentation'],
          'verifiableCredential': vcsResult
        },
        'exp': DateTime.now()
                .add(const Duration(days: 30))
                .millisecondsSinceEpoch ~/
            1000,
        'nonce': objectPayload['nonce'], // unique nonce
        'jti':
            'https://digitalwallet.moda:8443/vp/api/presentation/${uuid.v4()}'
      };

      String jwtToken = await _signJwt(header, payload);

      //Build Presentation Submission
      final presetationSubmission = {
        'id': uuid.v4(),
        'definition_id': objectPayload['presentation_definition']['id'],
        'descriptor_map': descriptorMap
      };

      String customJwt = "";
      if (customData != "") {
        Map<String, dynamic> customPayload = jsonDecode(customData);
        customJwt = await _signJwt(header, customPayload);
      }

      Map<String, dynamic> response = await httpService.sendVP(
          objectPayload['response_uri'],
          objectPayload['state'],
          jwtToken,
          jsonEncode(presetationSubmission),
          customJwt);
      if (response['error'] != null) {
        return utils.response(
            code: '4021', message: response['error'], data: jwtToken);
      }

      return utils.response(code: '0', message: 'SUCCESS');
    } catch (e) {
      return utils.response(code: '1', message: e.toString());
    }
  }

  /// generateVPNFC method
  /// This method generates VP for NFC transmission
  Future<String> generateVPNFC(String didFile, String vc) async {
    try {
      Map<String, dynamic> didMap = jsonDecode(didFile);

      final header = {
        'typ': 'JWT',
        'alg': 'ES256',
        "jwk": didMap['verificationMethod'][0]['publicKeyJwk']
      };

      const uuid = Uuid();
      final payload = {
        'sub': didMap['id'],
        'aud': didMap['id'],
        'iss': didMap['id'],
        'nbf': DateTime.now().millisecondsSinceEpoch ~/ 1000,
        'vp': {
          'context': ['https://www.w3.org/2018/credentials/v1'],
          'type': ['VerifiablePresentation'],
          'verifiableCredential': [vc]
        },
        'exp': DateTime.now()
                .add(const Duration(days: 30))
                .millisecondsSinceEpoch ~/
            1000,
        'nonce': uuid.v4(),
        'jti':
            'https://digitalwallet.moda:8443/vp/api/presentation/${uuid.v4()}'
      };

      String jwtToken = await _signJwt(header, payload);

      return utils.response(code: '0', message: jwtToken);
    } catch (e) {
      return utils.response(code: '1', message: e.toString());
    }
  }

  /// verifyVPNFC method
  /// This method verifies VP received via NFC
  Future<String> verifyVPNFC(String vp) async {
    try {
      var vpFlg = false;
      final vpPayload = utils.jwtDecode(vp);
      final vcToken = vpPayload['vp']['verifiableCredential'][0];
      final vcPayload = utils.jwtDecode(vcToken);

      print('xxxxxx');
      print(vcPayload);
      return utils.response(code: '0', message: 'SUCCESS');
    } catch (e) {
      return utils.response(code: '1', message: e.toString());
    }
  }

  /// checkVCValid method
  /// This method checks if VC is valid
  Future<bool> checkVCValid(Map vcStatus, Map vcListToken) async {
    final vcListPayload = utils.jwtDecode(vcListToken['statusList']);

    Uint8List decodedBytes =
        base64.decode(vcListPayload['vc']['credentialSubject']['encodedList']);
    // GZIP unzip
    Uint8List decompressedBytes =
        Uint8List.fromList(GZipDecoder().decodeBytes(decodedBytes));

    // transfer bytes to bits
    List<int> bitArray = utils.bytesToBits(decompressedBytes);
    if (bitArray[int.parse(vcStatus['statusListIndex'])] == 0) {
      return true;
    } else {
      return false;
    }
  }

  /// parsePresentationDefinition method
  /// This method parses presentation definition
  List parsePresentationDefinition(Map<String, dynamic> objectPayload) {
    final definition = objectPayload['presentation_definition'];
    var result = [];

    if (definition.containsKey('submission_requirements')) {
      for (var req in definition['submission_requirements']) {
        var request = {
          'name': req['name'],
          'group': req['from'],
          'rule': req['rule'],
          'count': req.containsKey('count') ? req['count'] : null,
          'max': req.containsKey('max') ? req['max'] : null,
          'cards': []
        };
        for (var vc in definition['input_descriptors']) {
          if (vc['group'].contains(req['from'])) {
            var card = {
              'card': vc['constraints']['fields'][0]['filter']['contains']
                  ['const'],
              'card_id': vc['id'],
              'name': vc['name'],
              'fields': []
            };
            for (int i = 1; i < vc['constraints']['fields'].length; i++) {
              final path = vc['constraints']['fields'][i]['path'][0];
              final parts = path.split('credentialSubject.');
              if (parts.length > 1) {
                card['fields'].add(parts[1]);
              }
            }
            request['cards'].add(card);
          }
        }
        result.add(request);
      }
    } else {
      var request = {'cards': []};
      for (var vc in definition['input_descriptors']) {
        var card = {
          'card': vc['constraints']['fields'][0]['filter']['contains']['const'],
          'name': vc['name'],
          'fields': []
        };
        for (int i = 1; i < vc['constraints']['fields'].length; i++) {
          final path = vc['constraints']['fields'][i]['path'][0];
          final parts = path.split('credentialSubject.');
          if (parts.length > 1) {
            card['fields'].add(parts[1]);
          }
        }
        request['cards']?.add(card);
      }
      result.add(request);
    }
    return result;
  }

  /// transferVC method
  /// This method transfers VC
  Future<String> transferVC(String didFile, String vc) async {
    try {
      Map<String, dynamic> didMap = jsonDecode(didFile);
      final objectPayload = utils.jwtDecode(vc);

      final header = {
        'typ': 'JWT',
        'alg': 'ES256',
        "jwk": didMap['verificationMethod'][0]['publicKeyJwk']
      };

      const uuid = Uuid();
      final payload = {
        'sub': didMap['id'],
        'aud': objectPayload['sub'],
        'iss': didMap['id'],
        'nbf': DateTime.now().millisecondsSinceEpoch ~/ 1000,
        'vp': {
          'context': ['https://www.w3.org/2018/credentials/v1'],
          'type': ['VerifiablePresentation'],
          'verifiableCredential': [vc]
        },
        'exp': DateTime.now()
                .add(const Duration(days: 30))
                .millisecondsSinceEpoch ~/
            1000,
        'nonce': objectPayload['nonce'],
        'jti':
            'https://digitalwallet.moda:8443/vp/api/presentation/${uuid.v4()}'
      };

      String jwtToken = await _signJwt(header, payload);

      String url = objectPayload['jti'];
      Map<String, dynamic> response =
          await httpService.transferVC(url, jwtToken);
      if (response['error'] != null) {
        return utils.response(
            code: '7011', message: response['error'], data: jwtToken);
      }

      return utils.response(code: '0', message: 'SUCCESS', data: response);
    } catch (e) {
      return utils.response(code: '1', message: e.toString());
    }
  }

  /// sendRequest method
  /// This method sends API request
  Future<String> sendRequest(String url, String type, String body) async {
    try {
      Map<String, dynamic> bodyMap = jsonDecode(body);
      Map<String, dynamic> response;
      response = await httpService.sendRequest(url, type, bodyMap);
      if (response['error'] != null) {
        return utils.response(code: '8011', message: response['error']);
      }

      return utils.response(code: '0', message: 'SUCCESS', data: response);
    } catch (e) {
      return utils.response(code: '1', message: e.toString());
    }
  }

  /// sendJWTRequest method
  /// This method sends JWT authenticated API request
  Future<String> sendJWTRequest(String url, String body, String didFile) async {
    try {
      Map<String, dynamic> didMap = jsonDecode(didFile);
      Map<String, dynamic> payload = jsonDecode(body);
      final header = {
        'typ': 'JWT',
        'alg': 'ES256',
        "kid": didMap['id']
      };

      String jwtToken = await _signJwt(header, payload);

      Map<String, dynamic> response =
          await httpService.sendRequest(url, 'POST', {'jwt': jwtToken});

      if (response['error'] != null) {
        return utils.response(code: '8012', message: response['error']);
      }

      return utils.response(
          code: '0', message: 'SUCCESS', data: response['response']);
    } catch (e) {
      return utils.response(code: '1', message: e.toString());
    }
  }

  Map<String, List<dynamic>> _vcListMap(List vcList) {
    final map = <String, List<dynamic>>{};
    for (var item in vcList) {
      if (item is Map<String, dynamic>) {
        item.forEach((key, value) {
          if (value is List) {
            map[key] = value;
          }
        });
      }
    }
    return map;
  }
}
