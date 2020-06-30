import 'dart:convert';

import 'package:http/http.dart' as http;
import 'dart:html' as html;

class BELLA {
  BELLA._();
  static final BELLA api = BELLA._();

  static final url = "http://localhost:9876/";
  String rawJWT;

  Future<bool> checkInitialized() async {
    var res = await http.get(url + "initialized");
    return res.statusCode != 200;
  }

  Future<bool> register(String name, String uniqueID, String password) async {
    var request = {
      "name": name,
      "unique_id": uniqueID,
      "password": password,
    };
    var res = await http.post(url + "auth/register", body: jsonEncode(request));
    return res.statusCode == 200;
  }

  Future<bool> initialize(
    String name,
    String uniqueID,
    String password,
    String electionName,
    DateTime start,
    DateTime end,
    int minCandidates,
    int maxCandidates,
  ) async {
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

  Map<String, String> headers() {
    return {
      "Content-Type": "application/json",
      "Authorization": "Bearer $rawJWT",
    };
  }

  Future<bool> addCandidate(Candidate c) async {
    var request = {
      "name": c.name,
      "presentation": c.presentation,
      "image": c.image ?? "",
    };
    var res = await http.post(url + "candidates/add",
        body: jsonEncode(request), headers: headers());
    return res.statusCode == 200;
  }

  Future<User> login(String uniqueID, String password) async {
    var request = {
      "unique_id": uniqueID,
      "password": password,
    };
    var res = await http.post(url + "auth/login", body: jsonEncode(request));
    if (res.statusCode != 200) {
      return null;
    }
    rawJWT = jsonDecode(res.body);
    var jwt = parseJwt(rawJWT);
    return User.fromJson(jwt);
  }

  Future<List<Candidate>> getCandidates() async {
    var r = await http.get(url + "candidates/get");
    var res = jsonDecode(r.body);
    List<Candidate> cands = [];
    res.forEach((x) {
      cands.add(Candidate.fromJson(x));
    });
    return cands;
  }

  Future<List<User>> getUnvalidatedUsers() async {
    var r = await http.get(url + "users/unvalidated/get", headers: headers());
    var res =
        (jsonDecode(r.body) as List)?.map((u) => User.fromJson(u))?.toList() ??
            [];
    return res;
  }

  downloadFile(int id) async {
    print("Downloading file $id");
    var res = await http.get(url + "users/files/download?id=$id",headers: headers());
    final blob = html.Blob([res.bodyBytes]);
    final downloadURL = html.Url.createObjectUrlFromBlob(blob);
    final anchor = html.document.createElement('a') as html.AnchorElement
      ..href = downloadURL
      ..style.display = 'none'
      ..download = 'asd.pdf'; // TODO real name
    html.document.body.children.add(anchor);

    anchor.click();

    html.document.body.children.remove(anchor);
    html.Url.revokeObjectUrl(downloadURL);
  }
}

class User {
  User({this.id, this.name, this.uniqueID, this.role});

  int id;
  String name;
  String uniqueID;
  String role;
  List<UserFile> files;

  // equivalents for backend on consts.go
  static final String roleNone = "none";
  static final String roleValidated = "validated";
  static final String roleAdmin = "admin";

  User.fromJson(Map<String, dynamic> json)
      : id = json["id"],
        name = json["name"],
        uniqueID = json["unique_id"],
        role = json["role"],
        files = (json["files"] as List)
                ?.map((r) => UserFile.fromJson(r))
                ?.toList() ??
            [];
}

class UserFile {
  UserFile({this.id, this.description});

  int id;
  String description;

  UserFile.fromJson(Map<String, dynamic> json)
      : id = json["id"],
        description = json["description"];
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
