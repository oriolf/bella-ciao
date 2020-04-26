import 'dart:convert';

import 'package:http/http.dart' as http;

class API {
  static final url = "http://localhost:9876/";

  static Future<bool> checkInitialized() async {
    var res = await http.get(url + "initialized");
    return res.statusCode != 200;
  }

  static Future<bool> register(
      String name, String uniqueID, String password) async {
    var request = {
      "name": name,
      "unique_id": uniqueID,
      "password": password,
    };
    var res = await http.post(url + "auth/register", body: jsonEncode(request));
    return res.statusCode == 200;
  }

  static Future<bool> initialize(
    String name,
    String uniqueID,
    String password,
    String electionName,
    DateTime start,
    DateTime end,
    int minCandidates,
    int maxCandidates,
  ) async {
    print("START: ${start.toUtc().toIso8601String()}");
    var request = {
      "admin": {
        "name": name,
        "unique_id": uniqueID,
        "password": password,
      },
      "election": {
        "name": electionName,
        "start": start.toUtc().toIso8601String(),
        "end": end.toUtc().toIso8601String(),
        // TODO allow selecting from backend-supplied list
        "count_type": "borda",
        "min_candidates": minCandidates,
        "max_candidates": maxCandidates,
      }
    };
    var res = await http.post(url + "initialize", body: jsonEncode(request));
    return res.statusCode == 200;
  }

  static Future<User> login(String uniqueID, String password) async {
    var request = {
      "unique_id": uniqueID,
      "password": password,
    };
    var res = await http.post(url + "auth/login", body: jsonEncode(request));
    if (res.statusCode != 200) {
      return null;
    }
    var jwt = parseJwt(jsonDecode(res.body));
    return User.fromJson(jwt);
  }

  static Future<List<Candidate>> getCandidates() async {
    var r = await http.get(url + "candidates/get");
    var res = jsonDecode(r.body);
    List<Candidate> cands = [];
    res.forEach((x) {
      cands.add(Candidate.fromJson(x));
    });
    return cands;
  }
}

class User {
  User({this.id, this.name, this.uniqueID, this.role});

  int id;
  String name;
  String uniqueID;
  String role;

  User.fromJson(Map<String, dynamic> json)
      : id = json["id"],
        name = json["name"],
        uniqueID = json["unique_id"],
        role = json["role"];
}

class Candidate {
  Candidate({this.id, this.name, this.presentation, this.image});

  int id;
  String name;
  String presentation;
  String image;

  Candidate.fromJson(Map<String, dynamic> json)
      : id = json["id"],
        name = json["name"],
        presentation = json["presentation"],
        image = json["image"];
}

Map<String, dynamic> parseJwt(String token) {
  final parts = token.split('.');
  if (parts.length != 3) {
    throw Exception('invalid token');
  }

  final payload = _decodeBase64(parts[1]);
  final payloadMap = json.decode(payload);
  if (payloadMap is! Map<String, dynamic>) {
    throw Exception('invalid payload');
  }

  return payloadMap;
}

String _decodeBase64(String str) {
  String output = str.replaceAll('-', '+').replaceAll('_', '/');

  switch (output.length % 4) {
    case 0:
      break;
    case 2:
      output += '==';
      break;
    case 3:
      output += '=';
      break;
    default:
      throw Exception('Illegal base64url string!"');
  }

  return utf8.decode(base64Url.decode(output));
}
