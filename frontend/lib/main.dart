import 'package:flutter/material.dart';
import 'package:bella_ciao/api.dart';

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
  @override
  Widget build(BuildContext context) {
    return Page(
      title: "Inici",
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          LoginForm(),
        ],
      ),
    );
  }
}

class LoginForm extends StatefulWidget {
  @override
  _LoginFormState createState() => _LoginFormState();
}

class _LoginFormState extends State<LoginForm> {
  final _formKey = GlobalKey<FormState>();
  final _idController = TextEditingController();
  final _nameController = TextEditingController();
  final _passwordController = TextEditingController();
  final _passwordConfirmController = TextEditingController();
  bool _registering = false;
  String _title = "Login";
  String _errorText = "";

  Widget _buildIdInput() {
    return _buildTextInput(_idController, "ID");
  }

  Widget _buildNameInput() {
    return _buildTextInput(_nameController, "Name");
  }

  Widget _buildTextInput(TextEditingController controller, String name) {
    return TextFormField(
        controller: controller,
        decoration: InputDecoration(
          border: OutlineInputBorder(),
          hintText: name,
        ),
        validator: (value) {
          if (value.length > 0) {
            return null;
          }
          return "$name is required";
        });
  }

  Widget _buildPasswordInput() {
    return TextFormField(
        controller: _passwordController,
        obscureText: true,
        decoration: InputDecoration(
          border: OutlineInputBorder(),
          hintText: "Password",
        ),
        validator: (value) {
          if (value.length > 4) {
            // TODO use the same as MIN_PASSWORD_LENGTH set in backend
            return null;
          }
          return "Password must be at least 4 characters long";
        });
  }

  Widget _buildPasswordConfirmInput() {
    return TextFormField(
        controller: _passwordConfirmController,
        obscureText: true,
        decoration: InputDecoration(
          border: OutlineInputBorder(),
          hintText: "Confirm password",
        ),
        validator: (value) {
          if (value == _passwordController.value.text) {
            return null;
          }
          return "Passwords must match";
        });
  }

  Widget _buildSubmitButton() {
    return FlatButton(
      child: Text(_title),
      color: Colors.blue,
      textColor: Colors.white,
      onPressed: _login,
    );
  }

  Widget _buildRegisterButton() {
    return FlatButton(
      child: Text("Register"),
      onPressed: () {
        setState(() {
          _registering = true;
          _title = "Register";
        });
      },
    );
  }

  _login() async {
    if (_formKey.currentState.validate()) {
      setState(() {
        _errorText = "";
      });
      if (_registering) {
        var res = await API.register(_nameController.value.text,
            _idController.value.text, _passwordController.value.text);

        if (!res) {
          setState(() {
            _errorText = "Register couldn't be completed";
          });
        }
      } else {
        var user = await API.login(
            _idController.value.text, _passwordController.value.text);

        if (user == null) {
          setState(() {
            _errorText = "Could not log in";
          });
        }
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    var children = <Widget>[
      Text(_title, style: Theme.of(context).textTheme.headline4),
      SizedBox(height: 10),
      _buildIdInput(),
      SizedBox(height: 10),
      _buildPasswordInput(),
      SizedBox(height: 10),
    ];
    if (_registering) {
      children.add(_buildPasswordConfirmInput());
      children.add(SizedBox(height: 10));
      children.add(_buildNameInput());
      children.add(Text(_errorText));
      children.add(_buildSubmitButton());
    } else {
      children.add(Text(_errorText));
      children.add(Row(children: <Widget>[
        _buildSubmitButton(),
        _buildRegisterButton(),
      ]));
    }
    return Card(
      child: Container(
        margin: EdgeInsets.all(20),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: children,
          ),
        ),
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
