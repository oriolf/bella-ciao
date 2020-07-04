import 'package:flutter/material.dart';

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
