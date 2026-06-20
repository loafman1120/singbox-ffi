import 'package:singbox_ffi/singbox_ffi.dart';

const configJson = '''
{
  "log": {"level": "info"},
  "inbounds": [
    {
      "type": "mixed",
      "tag": "mixed-in",
      "listen": "127.0.0.1",
      "listen_port": 2080
    }
  ],
  "outbounds": [
    {"type": "direct", "tag": "direct"}
  ],
  "route": {"final": "direct"}
}
''';

void main(List<String> args) {
  final libraryPath = args.isEmpty ? null : args.first;
  final singbox = SingboxFfi.openBundled(libraryPath);

  print('sing-box version: ${singbox.version()}');
  print('Go version: ${singbox.goVersion()}');

  singbox.init();
  singbox.checkConfig(configJson);

  final service = singbox.start(configJson);
  try {
    print('mixed proxy listening on 127.0.0.1:2080');
    print('service state: ${service.state().state.name}');
    for (final event in service.drainLogs()) {
      if (event.isReset) {
        print('logs reset');
      } else {
        print('[${event.levelName}] ${event.message}');
      }
    }
    service.reload(configJson);
  } finally {
    service.close();
  }
}
