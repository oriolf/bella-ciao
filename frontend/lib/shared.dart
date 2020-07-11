import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class NotificationBox extends StatelessWidget {
  NotificationBox({this.content, this.type, this.action});

  final String content;
  final String type;
  final Function action;

  @override
  Widget build(BuildContext context) {
    var color = Colors.blue;
    if (type == "error") {
      color = Colors.red;
    } else if (type == "warning") {
      color = Colors.orange;
    }
    return Container(
      decoration: BoxDecoration(
        color: color[100],
        border: Border.all(color: color[700]),
        borderRadius: BorderRadius.all(Radius.circular(5)),
      ),
      padding: EdgeInsets.all(15),
      child: Row(children: <Widget>[
        Expanded(flex: 5, child: Text(content)),
        action != null
            ? Expanded(
                flex: 1,
                child: FlatButton(
                    child: Text("Resol"),
                    color: Colors.white,
                    onPressed: action),
              )
            : Container(),
      ]),
    );
  }
}

class UpdateField extends StatefulWidget {
  UpdateField({this.updateFunc, this.name, this.original});

  final Function updateFunc;
  final String name;
  final String original;

  @override
  _UpdateFieldState createState() =>
      _UpdateFieldState(updateFunc: updateFunc, name: name, original: original);
}

class _UpdateFieldState extends State<UpdateField> {
  _UpdateFieldState({this.updateFunc, this.name, this.original});

  final Future<String> Function(String) updateFunc;
  final String name;
  final String original;
  TextEditingController _controller;
  bool _updating = false;

  void initState() {
    super.initState();
    _controller = TextEditingController(text: original);
  }

  Widget _buildTextInput() {
    return TextField(
        controller: _controller,
        hintText: name,
        validator: (value) {
          if (value.length > 0) {
            return null;
          }
          return "$name is required";
        });
  }

// TODO _updating == true show loading icon inside button
  _update() async {
    setState(() {
      _updating = true;
    });
    var error = await updateFunc(_controller.value.text);
    if (error != "") {
      Scaffold.of(context).showSnackBar(SnackBar(
        content: Text(error),
        duration: Duration(seconds: 3),
      ));
    }
    setState(() {
      _updating = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        Expanded(child: _buildTextInput(), flex: 1),
        Expanded(
            child: FlatButton(
              child: Text("Actualitza"),
              color: Colors.blue,
              textColor: Colors.white,
              onPressed: _update,
            ),
            flex: 1),
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
