import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';

void main() {
  runApp(BellaCiao());
}

class BellaCiao extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Bella Ciao',
      theme: ThemeData(
        primarySwatch: Colors.blue,
        visualDensity: VisualDensity.adaptivePlatformDensity,
      ),
      home: HomePage(),
    );
  }
}

class Page extends StatelessWidget {
  Page({this.title, this.body});

  final String title;
  final Widget body;

  Function _navigate(BuildContext context, Function builder) {
    return () {
      Navigator.of(context).pop();
      Navigator.of(context).push(MaterialPageRoute(builder: builder));
    };
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(title),
      ),
      drawer: Drawer(
        child: ListView(children: <Widget>[
          ListTile(
            leading: Icon(Icons.home),
            title: Text("Inici"),
            onTap: _navigate(context, (BuildContext context) => HomePage()),
          ),
          ListTile(
            leading: Icon(Icons.question_answer),
            title: Text("Preguntes freqüents"),
            onTap: _navigate(context, (BuildContext context) => FAQPage()),
          ),
          ListTile(
            leading: Icon(Icons.people),
            title: Text("Candidatures"),
            onTap:
                _navigate(context, (BuildContext context) => CandidatesPage()),
          ),
        ]),
      ),
      body: SingleChildScrollView(
        child: Container(
          margin: EdgeInsets.all(40),
          child: body,
        ),
      ),
    );
  }
}

class HomePage extends StatelessWidget {
  final TextEditingController _idController = TextEditingController();
  final TextEditingController _passwordController = TextEditingController();

  _login() {
    print("Logging in with ID ${_idController.value.text} and password ${_passwordController.value.text}");
  }

  @override
  Widget build(BuildContext context) {
    var _idInput = TextField(
      decoration: InputDecoration(
        border: OutlineInputBorder(),
        hintText: "ID",
      ),
      controller: _idController,
    );
    var _passwordInput = TextField(
      obscureText: true,
      decoration: InputDecoration(
        border: OutlineInputBorder(),
        hintText: "Password",
      ),
      controller: _passwordController,
    );
    var _submitButton = FlatButton(child: Text("Log in"), color: Colors.blue, textColor: Colors.white, onPressed: _login);
    return Page(
      title: "Inici",
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Card(
            child: Container(
              margin: EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text("Login", style: Theme.of(context).textTheme.headline4),
                  SizedBox(height: 10),
                  _idInput,
                  SizedBox(height: 10),
                  _passwordInput,
                  SizedBox(height: 10),
                  _submitButton,
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class FAQ {
  FAQ({this.question, this.answer});
  final String question, answer;
}

class FAQPage extends StatelessWidget {
  final List<FAQ> qas = [
    FAQ(question: "Question one", answer: "Answer one"),
    FAQ(question: "Question two", answer: "Answer two"),
    FAQ(question: "Question three", answer: "Answer three"),
    FAQ(question: "Question four", answer: "Answer four"),
  ];

  Widget _faq(BuildContext context, FAQ f) {
    return Container(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(f.question, style: Theme.of(context).textTheme.headline4),
          Text(f.answer),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Page(
      title: "FAQ",
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: qas.map((x) => _faq(context, x)).toList(),
      ),
    );
  }
}

class CandidatesPage extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Page(
        title: "Candidatures", body: Center(child: Text("Candidatures")));
  }
}
