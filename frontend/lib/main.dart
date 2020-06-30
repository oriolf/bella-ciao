import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:bella_ciao/api.dart';
import 'package:intl/intl.dart';
import 'package:datetime_picker_formfield/datetime_picker_formfield.dart';

const ROLE_ADMIN = "admin";

void main() {
  runApp(BellaCiao());
}

class BellaCiao extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => JWT()),
      ],
      child: Consumer<JWT>(builder: (context, jwt, _) {
        return MaterialApp(
          title: 'Bella Ciao',
          theme: ThemeData(
            primarySwatch: Colors.blue,
            visualDensity: VisualDensity.adaptivePlatformDensity,
          ),
          home: Stack(
            children: <Widget>[
              HomePage(jwt: jwt),
              CheckInitialized(),
            ],
          ),
        );
      }),
    );
  }
}

class CheckInitialized extends StatefulWidget {
  @override
  _CheckInitializedState createState() => _CheckInitializedState();
}

class _CheckInitializedState extends State<CheckInitialized> {
  bool _alreadyChecked = false;

  _checkInitialized(BuildContext context) async {
    setState(() {
      _alreadyChecked = true;
    });
    var jwt = Provider.of<JWT>(context, listen: false);
    var initialized = await BELLA.api.checkInitialized();
    if (!initialized) {
      Navigator.of(context).pop();
      Navigator.of(context).push(
        MaterialPageRoute(
          builder: (context) => ChangeNotifierProvider.value(
            value: jwt,
            child: InitializePage(jwt: jwt),
          ),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    if (!_alreadyChecked) {
      _checkInitialized(context);
    }

    return Container();
  }
}

class Page extends StatelessWidget {
  Page({this.title, this.body, this.jwt, this.floatingActionButton});

  final String title;
  final Widget body;
  final JWT jwt;
  final Widget floatingActionButton;

  Function _navigate(BuildContext context, Function builder) {
    return () {
      Navigator.of(context).pop();
      Navigator.of(context).push(MaterialPageRoute(
          builder: (context) => ChangeNotifierProvider.value(
                value: jwt,
                child: builder(context),
              )));
    };
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(title + (jwt.user != null ? " logged" : "")),
      ),
      drawer: Drawer(
        child: ListView(children: <Widget>[
          ListTile(
            leading: Icon(Icons.home),
            title: Text("Inici"),
            onTap: _navigate(
                context, (BuildContext context) => HomePage(jwt: jwt)),
          ),
          ListTile(
            leading: Icon(Icons.question_answer),
            title: Text("Preguntes freqüents"),
            onTap:
                _navigate(context, (BuildContext context) => FAQPage(jwt: jwt)),
          ),
          ListTile(
            leading: Icon(Icons.people),
            title: Text("Candidatures"),
            onTap: _navigate(
                context, (BuildContext context) => CandidatesPage(jwt: jwt)),
          ),
          if (jwt.user != null)
            ListTile(
                leading: Icon(Icons.exit_to_app),
                title: Text("Surt"),
                onTap: () {
                  Provider.of<JWT>(context, listen: false).invalidateUser();
                  Navigator.of(context).pop();
                }),
        ]),
      ),
      body: SingleChildScrollView(
        child: Container(
          margin: EdgeInsets.all(40),
          child: body,
        ),
      ),
      floatingActionButton: floatingActionButton,
    );
  }
}

class HomePage extends StatelessWidget {
  HomePage({this.jwt});

  final JWT jwt;

  @override
  Widget build(BuildContext context) {
    return Page(
      jwt: jwt,
      title: "Inici",
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          if (jwt.user == null) LoginForm(),
          if (jwt.user != null) HomePageContent(jwt: jwt),
        ],
      ),
    );
  }
}

class HomePageContent extends StatelessWidget {
  HomePageContent({this.jwt});

  final JWT jwt;

  @override
  Widget build(BuildContext context) {
    if (jwt.user.role == User.roleNone) {
      return HomePageContentNone(jwt: jwt);
    } else if (jwt.user.role == User.roleValidated) {
      return HomePageContentValidated(jwt: jwt);
    } else if (jwt.user.role == User.roleAdmin) {
      return HomePageContentAdmin(jwt: jwt);
    }
    return Center(child: Text("Unknown user role!"));
  }
}

class HomePageContentNone extends StatelessWidget {
  HomePageContentNone({this.jwt});

  final JWT jwt;

// TODO allow to upload or erase documentation, and mark rejection comments as resolved
  @override
  Widget build(BuildContext context) {
    return Center(child: Text("You are logged in, but not validated!"));
  }
}

class HomePageContentValidated extends StatelessWidget {
  HomePageContentValidated({this.jwt});

  final JWT jwt;

// TODO show basic election info, and if time is right, allow to vote directly here
  @override
  Widget build(BuildContext context) {
    return Center(child: Text("You are logged in, and validated!"));
  }
}

class HomePageContentAdmin extends StatelessWidget {
  HomePageContentAdmin({this.jwt});

  final JWT jwt;

// TODO implement, show user and files info, also buttons to validate and to reject (with a message)
  Widget _buildUnvalidatedUser(User x) {
    var children = <Widget>[Text("${x.name} (${x.uniqueID})")];
    for (var file in x.files) {
      children.add(Row(children: <Widget>[
        OutlineButton(
          visualDensity: VisualDensity.compact,
          child: Text("Visualize"),
          onPressed: () { BELLA.api.downloadFile(file.id); },
        ),
        SizedBox(width: 20),
        Text("${file.description}")
      ]));
    }
    return ListTile(
      title: Column(
          crossAxisAlignment: CrossAxisAlignment.start, children: children),
      trailing: Container(
        width: 200,
        child: Row(children: <Widget>[
          FlatButton(
            child: Text("Validate"),
            color: Colors.blue,
            textColor: Colors.white,
            onPressed: () {},
          ),
          SizedBox(width: 20),
          FlatButton(
            child: Text("Reject"),
            color: Colors.red,
            textColor: Colors.white,
            onPressed: () {},
          ),
        ]),
      ),
    );
  }

// TODO show list of unvalidated users with button to validate or reject with a message
// show first users without rejection message
  @override
  Widget build(BuildContext context) {
    return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text("Users pending validation", style: TextStyle(fontSize: 36)),
          Container(
            decoration: BoxDecoration(border: Border.all(color: Colors.grey)),
            height: 500,
            child: AsyncList(
              builder: _buildUnvalidatedUser,
              dataFunc: BELLA.api.getUnvalidatedUsers,
            ),
          )
        ]);
  }
}

class AsyncList<T> extends StatefulWidget {
  AsyncList({this.builder, this.dataFunc});

  final Widget Function(T) builder;
  final Future<List<T>> Function() dataFunc;

  @override
  _AsyncListState<T> createState() =>
      _AsyncListState<T>(builder: builder, dataFunc: dataFunc);
}

class _AsyncListState<T> extends State<AsyncList<T>> {
  _AsyncListState({this.builder, this.dataFunc});

  final Widget Function(T) builder;
  final Future<List<T>> Function() dataFunc;
  List<T> data = [];
  bool waiting = true;

  void initState() {
    super.initState();
    reloadData();
  }

  reloadData() async {
    var l = await dataFunc();
    setState(() {
      data = l;
      waiting = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    if (waiting) {
      return Center(child: Text("waiting")); // TODO loader
    }
    return ListView(
      children: data.map((x) => builder(x)).toList(),
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
    return TextField(
        controller: controller,
        hintText: name,
        validator: (value) {
          if (value.length > 0) {
            return null;
          }
          return "$name is required";
        });
  }

  Widget _buildPasswordInput() {
    return TextField(
        controller: _passwordController,
        obscureText: true,
        hintText: "Password",
        validator: (value) {
          if (value.length >= 4) {
            // TODO use the same as MIN_PASSWORD_LENGTH set in backend
            return null;
          }
          return "Password must be at least 4 characters long";
        });
  }

  Widget _buildPasswordConfirmInput() {
    return TextField(
        controller: _passwordConfirmController,
        obscureText: true,
        hintText: "Confirm password",
        validator: (value) {
          if (value == _passwordController.value.text) {
            return null;
          }
          return "Passwords must match";
        });
  }

  Widget _buildSubmitButton(BuildContext context) {
    return FlatButton(
      child: Text(_title),
      color: Colors.blue,
      textColor: Colors.white,
      onPressed: () {
        _login(context);
      },
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

  _login(BuildContext context) async {
    if (_formKey.currentState.validate()) {
      setState(() {
        _errorText = "";
      });
      if (_registering) {
        var res = await BELLA.api.register(_nameController.value.text,
            _idController.value.text, _passwordController.value.text);

        if (!res) {
          setState(() {
            _errorText = "Register couldn't be completed";
          });
        }
      } else {
        var user = await BELLA.api
            .login(_idController.value.text, _passwordController.value.text);

        var jwt = Provider.of<JWT>(context, listen: false);
        if (user == null) {
          setState(() {
            _errorText = "Could not log in";
          });
          jwt.invalidateUser();
        } else {
          jwt.updateUser(user);
        }
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Container(
        margin: EdgeInsets.all(20),
        child: Form(
          key: _formKey,
          child: SpacedColumn(
            padding: 10,
            children: <Widget>[
              Text(_title, style: Theme.of(context).textTheme.headline4),
              _buildIdInput(),
              _buildPasswordInput(),
              if (_registering) _buildPasswordConfirmInput(),
              if (_registering) _buildNameInput(),
              if (_registering) Text(_errorText),
              if (_registering) _buildSubmitButton(context),
              if (!_registering) Text(_errorText),
              if (!_registering)
                Row(children: <Widget>[
                  _buildSubmitButton(context),
                  _buildRegisterButton(),
                ]),
            ],
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
  FAQPage({this.jwt});

  final JWT jwt;
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
      jwt: jwt,
      title: "FAQ",
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: qas.map((x) => _faq(context, x)).toList(),
      ),
    );
  }
}

class CandidatesPage extends StatefulWidget {
  CandidatesPage({this.jwt});
  final JWT jwt;

  @override
  _CandidatesPageState createState() => _CandidatesPageState(jwt: jwt);
}

class _CandidatesPageState extends State<CandidatesPage> {
  _CandidatesPageState({this.jwt});

  final JWT jwt;
  List<Candidate> candidates = [];
  TextEditingController nameController = new TextEditingController();
  TextEditingController presentationController = new TextEditingController();
  final GlobalKey<ScaffoldState> _scaffoldKey = new GlobalKey<ScaffoldState>();

  void initState() {
    super.initState();
    _getCandidates();
  }

  _getCandidates() async {
    var cands = await BELLA.api.getCandidates();
    setState(() {
      candidates = cands;
    });
  }

  _initializeForm() {
    setState(() {
      nameController.value =
          new TextEditingController.fromValue(new TextEditingValue(text: ""))
              .value;
      presentationController.value =
          new TextEditingController.fromValue(new TextEditingValue(text: ""))
              .value;
    });
  }

  _addCandidate() async {
    _initializeForm();
    final c = await showDialog<Candidate>(
      context: context,
      builder: (BuildContext context) {
        return NewCandidateDialog(
            nameController: nameController,
            presentationController: presentationController);
      },
    );

    if (c != null) {
      try {
        var res = await BELLA.api.addCandidate(c);
        if (!res) {
          throw "HTTP ERROR";
        }
      } catch (e) {
        _showError('Error afegint candidat: $e');
      }
      _getCandidates();
    }
  }

  _showError(String message) {
    _scaffoldKey.currentState.showSnackBar(SnackBar(
      content: Text(message),
      duration: Duration(seconds: 3),
    ));
  }

  Widget _candidate(Candidate c) {
    return Container(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(c.name, style: Theme.of(context).textTheme.headline4),
          Text(c.presentation),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    Widget action;
    if (jwt.user != null && jwt.user.role == ROLE_ADMIN) {
      action = FloatingActionButton(
        child: Icon(Icons.add),
        onPressed: _addCandidate,
      );
    }
    return Page(
      jwt: jwt,
      title: "Candidatures",
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: candidates.map((x) => _candidate(x)).toList(),
      ),
      floatingActionButton: action,
    );
  }
}

class InitializePage extends StatelessWidget {
  InitializePage({this.jwt});

  final JWT jwt;

  @override
  Widget build(BuildContext context) {
    return Page(
      jwt: jwt,
      title: "Inicialitza",
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          InitializeForm(),
        ],
      ),
    );
  }
}

class InitializeForm extends StatefulWidget {
  @override
  _InitializeFormState createState() => _InitializeFormState();
}

class _InitializeFormState extends State<InitializeForm> {
  final _formKey = GlobalKey<FormState>();
  final _idController = TextEditingController();
  final _nameController = TextEditingController();
  final _passwordController = TextEditingController();
  final _passwordConfirmController = TextEditingController();
  final _electionNameController = TextEditingController();
  final _minCandidatesController = TextEditingController();
  final _maxCandidatesController = TextEditingController();
  DateTime _start;
  DateTime _end;
  String _errorText = "";

  Widget _buildIdInput() {
    return _buildTextInput(_idController, "ID");
  }

  Widget _buildNameInput() {
    return _buildTextInput(_nameController, "Name");
  }

  Widget _buildElectionNameInput() {
    return _buildTextInput(_electionNameController, "Name");
  }

  Widget _buildTextInput(TextEditingController controller, String name) {
    return TextField(
        controller: controller,
        hintText: name,
        validator: (value) {
          if (value.length > 0) {
            return null;
          }
          return "$name is required";
        });
  }

  Widget _buildMinCandidatesInput() {
    return _buildNumericInput(
        _minCandidatesController, "Minimum of candidates");
  }

  // TODO required max >= min
  Widget _buildMaxCandidatesInput() {
    return _buildNumericInput(
        _maxCandidatesController, "Maximum of candidates");
  }

  Widget _buildNumericInput(TextEditingController controller, String name) {
    return TextField(
        controller: controller,
        keyboardType: TextInputType.number,
        inputFormatters: <TextInputFormatter>[
          WhitelistingTextInputFormatter.digitsOnly
        ],
        hintText: name,
        validator: (value) {
          if (value.length > 0) {
            return null;
          }
          return "$name is required";
        });
  }

  Widget _buildPasswordInput() {
    return TextField(
        controller: _passwordController,
        obscureText: true,
        hintText: "Password",
        validator: (value) {
          if (value.length > 4) {
            // TODO use the same as MIN_PASSWORD_LENGTH set in backend
            return null;
          }
          return "Password must be at least 4 characters long";
        });
  }

  Widget _buildPasswordConfirmInput() {
    return TextField(
        controller: _passwordConfirmController,
        obscureText: true,
        hintText: "Confirm password",
        validator: (value) {
          if (value == _passwordController.value.text) {
            return null;
          }
          return "Passwords must match";
        });
  }

  Widget _buildStartInput() {
    return _buildDateInput("Start date", (val) {
      setState(() {
        _start = val;
      });
    });
  }

  Widget _buildEndInput() {
    return _buildDateInput("End date", (val) {
      setState(() {
        _end = val;
      });
    });
  }

  Widget _buildDateInput(String name, Function f) {
    return DateTimeField(
      validator: (value) {
        if (value != null) {
          return null;
        }
        return "$name is required";
      },
      decoration: InputDecoration(
        border: OutlineInputBorder(),
        hintText: "$name",
      ),
      format: DateFormat("yyyy-MM-dd HH:mm"),
      onShowPicker: (context, currentValue) async {
        final date = await showDatePicker(
            context: context,
            firstDate: DateTime(1900),
            initialDate: currentValue ?? DateTime.now(),
            lastDate: DateTime(2100));
        DateTime value;
        if (date != null) {
          final time = await showTimePicker(
            context: context,
            initialTime: TimeOfDay.fromDateTime(currentValue ?? DateTime.now()),
          );
          value = DateTimeField.combine(date, time);
        } else {
          value = currentValue;
        }
        f(value);
        return value;
      },
    );
  }

  Widget _buildSubmitButton(BuildContext context) {
    return FlatButton(
      child: Text("Initialize"),
      color: Colors.blue,
      textColor: Colors.white,
      onPressed: () {
        _initialize(context);
      },
    );
  }

  _initialize(BuildContext context) async {
    if (_formKey.currentState.validate()) {
      setState(() {
        _errorText = "";
      });
      var res = await BELLA.api.initialize(
        _nameController.value.text,
        _idController.value.text,
        _passwordController.value.text,
        _electionNameController.value.text,
        _start,
        _end,
        int.parse(_minCandidatesController.value.text),
        int.parse(_maxCandidatesController.value.text),
      );

      if (!res) {
        setState(() {
          _errorText = "Could not initialize";
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Form(
      key: _formKey,
      child: SpacedColumn(
        padding: 10,
        children: <Widget>[
          Text("Initialize", style: Theme.of(context).textTheme.headline4),
          Text("Admin user data", style: Theme.of(context).textTheme.headline5),
          _buildIdInput(),
          _buildPasswordInput(),
          _buildPasswordConfirmInput(),
          _buildNameInput(),
          Text("Election data", style: Theme.of(context).textTheme.headline6),
          _buildElectionNameInput(),
          _buildStartInput(),
          _buildEndInput(),
          _buildMinCandidatesInput(),
          _buildMaxCandidatesInput(),
          Text(_errorText),
          _buildSubmitButton(context),
        ],
      ),
    );
  }
}

class SpacedColumn extends StatelessWidget {
  SpacedColumn({this.children, this.padding});

  final List<Widget> children;
  final double padding;

  @override
  Widget build(BuildContext context) {
    var _children = <Widget>[];
    for (var child in children) {
      _children.add(child);
      _children.add(SizedBox(height: padding));
    }
    return Column(
        crossAxisAlignment: CrossAxisAlignment.start, children: _children);
  }
}

class NewCandidateDialog extends StatelessWidget {
  NewCandidateDialog({this.nameController, this.presentationController});

  final TextEditingController nameController;
  final TextEditingController presentationController;

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text("Add candidate"),
      content: Container(
        width: 500.0,
        //  may overflow, add scroll
        child: SpacedColumn(
          padding: 10,
          children: <Widget>[
            TextField(controller: nameController, hintText: "Name"),
            TextField(
              controller: presentationController,
              hintText: "Presentation",
              keyboardType: TextInputType.multiline,
              maxLines: 20,
            ),
          ],
        ),
      ),
      actions: <Widget>[
        FlatButton(
          child: Text('Cancel'),
          onPressed: () {
            Navigator.of(context).pop();
          },
        ),
        FlatButton(
          child: Text("Add"),
          onPressed: () {
            final c = new Candidate(
                name: nameController.value.text,
                presentation: presentationController.value.text);
            nameController.clear();
            presentationController.clear();

            Navigator.of(context).pop(c);
          },
        ),
      ],
    );
  }
}

class TextField extends StatelessWidget {
  const TextField({
    Key key,
    this.hintText,
    this.keyboardType,
    this.maxLines,
    this.validator,
    this.obscureText,
    this.inputFormatters,
    @required this.controller,
  }) : super(key: key);

  final TextEditingController controller;
  final String hintText;
  final int maxLines;
  final TextInputType keyboardType;
  final Function validator;
  final bool obscureText;
  final List<TextInputFormatter> inputFormatters;

  @override
  Widget build(BuildContext context) {
    return TextFormField(
      controller: controller,
      decoration: InputDecoration(
        border: OutlineInputBorder(),
        hintText: hintText,
      ),
      keyboardType: keyboardType,
      maxLines: maxLines ?? 1,
      validator: validator,
      obscureText: obscureText ?? false,
      inputFormatters: inputFormatters,
    );
  }
}

class JWT with ChangeNotifier {
  User _user;

  User get user => _user;

  updateUser(User u) {
    _user = u;
    notifyListeners();
  }

  invalidateUser() {
    _user = null;
    notifyListeners();
  }
}
