import 'package:singbox_ffi_flutter_example/singbox_ffi.dart';

const _configJson = '''
{
  "log": {
    "level": "info"
  },
  "inbounds": [
    {
      "type": "mixed",
      "tag": "mixed-in",
      "listen": "127.0.0.1",
      "listen_port": 2080
    }
  ],
  "outbounds": [
    {
      "type": "direct",
      "tag": "direct"
    }
  ],
  "route": {
    "final": "direct"
  }
}
''';

void main(List<String> args) {
  final libraryPath = args.isEmpty ? null : args.first;
  final singbox = SingboxFfi.open(libraryPath);

  print('sing-box version: ${singbox.version()}');
  print('go version: ${singbox.goVersion()}');

  singbox.init();
  singbox.checkConfig(_configJson);

  print('config is valid');
}
