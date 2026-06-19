import 'dart:async';
import 'dart:io';

import 'package:singbox_ffi_example/singbox_ffi.dart';

const _listen = '127.0.0.1';
const _port = 2080;

const _configJson = '''
{
  "log": {
    "level": "info"
  },
  "inbounds": [
    {
      "type": "mixed",
      "tag": "mixed-in",
      "listen": "$_listen",
      "listen_port": $_port
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

Future<void> main(List<String> args) async {
  final libraryPath = args.isEmpty ? null : args.first;
  final singbox = SingboxFfi.open(libraryPath);

  print('sing-box version: ${singbox.version()}');
  print('go version: ${singbox.goVersion()}');

  singbox.init();
  final service = singbox.start(_configJson);

  print('mixed proxy listening on $_listen:$_port');
  print('try: curl.exe -x socks5h://$_listen:$_port https://example.com');
  print('press Ctrl+C to stop');

  final stop = Completer<void>();
  late final StreamSubscription<ProcessSignal> sigint;
  sigint = ProcessSignal.sigint.watch().listen((_) {
    if (!stop.isCompleted) {
      stop.complete();
    }
  });

  if (!Platform.isWindows) {
    late final StreamSubscription<ProcessSignal> sigterm;
    sigterm = ProcessSignal.sigterm.watch().listen((_) {
      if (!stop.isCompleted) {
        stop.complete();
      }
    });
    await stop.future;
    await sigterm.cancel();
  } else {
    await stop.future;
  }

  await sigint.cancel();
  service.close();
  print('stopped');
}
