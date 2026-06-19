import 'dart:ffi';
import 'dart:io';

import 'package:ffi/ffi.dart';

final class SbInitOptions extends Struct {
  external Pointer<Utf8> basePath;
  external Pointer<Utf8> workingPath;
  external Pointer<Utf8> tempPath;
  external Pointer<Utf8> locale;
  external Pointer<Utf8> commandSecret;

  @Int32()
  external int commandPort;

  @Int32()
  external int logMaxLines;

  @Bool()
  external bool debug;

  @Bool()
  external bool oomKillerEnabled;

  @Bool()
  external bool oomKillerDisabled;

  @Int64()
  external int oomMemoryLimit;
}

typedef _SbVersionNative = Pointer<Utf8> Function();
typedef _SbVersionDart = Pointer<Utf8> Function();

typedef _SbGoVersionNative = Pointer<Utf8> Function();
typedef _SbGoVersionDart = Pointer<Utf8> Function();

typedef _SbFreeStringNative = Void Function(Pointer<Utf8>);
typedef _SbFreeStringDart = void Function(Pointer<Utf8>);

typedef _SbInitNative = Int32 Function(
  Pointer<SbInitOptions>,
  Pointer<Pointer<Utf8>>,
);
typedef _SbInitDart = int Function(
  Pointer<SbInitOptions>,
  Pointer<Pointer<Utf8>>,
);

typedef _SbCheckConfigNative = Int32 Function(
  Pointer<Utf8>,
  Pointer<Pointer<Utf8>>,
);
typedef _SbCheckConfigDart = int Function(
  Pointer<Utf8>,
  Pointer<Pointer<Utf8>>,
);

typedef _SbStartNative = Int32 Function(
  Pointer<Utf8>,
  Pointer<Uint64>,
  Pointer<Pointer<Utf8>>,
);
typedef _SbStartDart = int Function(
  Pointer<Utf8>,
  Pointer<Uint64>,
  Pointer<Pointer<Utf8>>,
);

typedef _SbReloadNative = Int32 Function(
  Uint64,
  Pointer<Utf8>,
  Pointer<Pointer<Utf8>>,
);
typedef _SbReloadDart = int Function(
  int,
  Pointer<Utf8>,
  Pointer<Pointer<Utf8>>,
);

typedef _SbStopNative = Int32 Function(Uint64, Pointer<Pointer<Utf8>>);
typedef _SbStopDart = int Function(int, Pointer<Pointer<Utf8>>);

typedef _SbFreeHandleNative = Int32 Function(Uint64);
typedef _SbFreeHandleDart = int Function(int);

class SingboxException implements Exception {
  SingboxException(this.message);

  final String message;

  @override
  String toString() => 'SingboxException: $message';
}

class SingboxInitOptions {
  const SingboxInitOptions({
    this.basePath = '.',
    this.workingPath = '.',
    this.tempPath = '.',
    this.locale,
    this.commandSecret = 'example-secret',
    this.commandPort = 0,
    this.logMaxLines = 300,
    this.debug = false,
    this.oomKillerEnabled = false,
    this.oomKillerDisabled = true,
    this.oomMemoryLimit = 0,
  });

  final String basePath;
  final String workingPath;
  final String tempPath;
  final String? locale;
  final String commandSecret;
  final int commandPort;
  final int logMaxLines;
  final bool debug;
  final bool oomKillerEnabled;
  final bool oomKillerDisabled;
  final int oomMemoryLimit;
}

class SingboxFfi {
  SingboxFfi._(this._lib)
      : _sbVersion = _lib.lookupFunction<_SbVersionNative, _SbVersionDart>(
          'sb_version',
        ),
        _sbGoVersion =
            _lib.lookupFunction<_SbGoVersionNative, _SbGoVersionDart>(
          'sb_go_version',
        ),
        _sbFreeString =
            _lib.lookupFunction<_SbFreeStringNative, _SbFreeStringDart>(
          'sb_free_string',
        ),
        _sbInit = _lib.lookupFunction<_SbInitNative, _SbInitDart>('sb_init'),
        _sbCheckConfig =
            _lib.lookupFunction<_SbCheckConfigNative, _SbCheckConfigDart>(
          'sb_check_config',
        ),
        _sbStart = _lib.lookupFunction<_SbStartNative, _SbStartDart>(
          'sb_start',
        ),
        _sbReload = _lib.lookupFunction<_SbReloadNative, _SbReloadDart>(
          'sb_reload',
        ),
        _sbStop = _lib.lookupFunction<_SbStopNative, _SbStopDart>(
          'sb_stop',
        ),
        _sbFreeHandle =
            _lib.lookupFunction<_SbFreeHandleNative, _SbFreeHandleDart>(
          'sb_free_handle',
        );

  factory SingboxFfi.open([String? path]) {
    return SingboxFfi._(DynamicLibrary.open(path ?? defaultLibraryName));
  }

  static String get defaultLibraryName {
    if (Platform.isWindows) {
      return 'lithenetcore.dll';
    }
    if (Platform.isMacOS || Platform.isIOS) {
      return 'liblithenetcore.dylib';
    }
    return 'liblithenetcore.so';
  }

  final DynamicLibrary _lib;
  final _SbVersionDart _sbVersion;
  final _SbGoVersionDart _sbGoVersion;
  final _SbFreeStringDart _sbFreeString;
  final _SbInitDart _sbInit;
  final _SbCheckConfigDart _sbCheckConfig;
  final _SbStartDart _sbStart;
  final _SbReloadDart _sbReload;
  final _SbStopDart _sbStop;
  final _SbFreeHandleDart _sbFreeHandle;

  String version() => _takeString(_sbVersion());

  String goVersion() => _takeString(_sbGoVersion());

  void init([SingboxInitOptions options = const SingboxInitOptions()]) {
    final opts = calloc<SbInitOptions>();
    final errOut = calloc<Pointer<Utf8>>();
    final allocations = <Pointer<Utf8>>[];

    Pointer<Utf8> nativeString(String? value) {
      if (value == null) {
        return nullptr;
      }
      final pointer = value.toNativeUtf8(allocator: calloc);
      allocations.add(pointer);
      return pointer;
    }

    try {
      opts.ref
        ..basePath = nativeString(options.basePath)
        ..workingPath = nativeString(options.workingPath)
        ..tempPath = nativeString(options.tempPath)
        ..locale = nativeString(options.locale)
        ..commandSecret = nativeString(options.commandSecret)
        ..commandPort = options.commandPort
        ..logMaxLines = options.logMaxLines
        ..debug = options.debug
        ..oomKillerEnabled = options.oomKillerEnabled
        ..oomKillerDisabled = options.oomKillerDisabled
        ..oomMemoryLimit = options.oomMemoryLimit;

      final code = _sbInit(opts, errOut);
      if (code != 0) {
        throw SingboxException(_takeError(errOut));
      }
    } finally {
      for (final pointer in allocations) {
        calloc.free(pointer);
      }
      calloc.free(errOut);
      calloc.free(opts);
    }
  }

  void checkConfig(String configJson) {
    final config = configJson.toNativeUtf8(allocator: calloc);
    final errOut = calloc<Pointer<Utf8>>();
    try {
      final code = _sbCheckConfig(config, errOut);
      if (code != 0) {
        throw SingboxException(_takeError(errOut));
      }
    } finally {
      calloc.free(config);
      calloc.free(errOut);
    }
  }

  SingboxService start(String configJson) {
    final config = configJson.toNativeUtf8(allocator: calloc);
    final handleOut = calloc<Uint64>();
    final errOut = calloc<Pointer<Utf8>>();
    try {
      final code = _sbStart(config, handleOut, errOut);
      if (code != 0) {
        throw SingboxException(_takeError(errOut));
      }
      return SingboxService._(this, handleOut.value);
    } finally {
      calloc.free(config);
      calloc.free(handleOut);
      calloc.free(errOut);
    }
  }

  void reload(int handle, String configJson) {
    final config = configJson.toNativeUtf8(allocator: calloc);
    final errOut = calloc<Pointer<Utf8>>();
    try {
      final code = _sbReload(handle, config, errOut);
      if (code != 0) {
        throw SingboxException(_takeError(errOut));
      }
    } finally {
      calloc.free(config);
      calloc.free(errOut);
    }
  }

  void stop(int handle) {
    final errOut = calloc<Pointer<Utf8>>();
    try {
      final code = _sbStop(handle, errOut);
      if (code != 0) {
        throw SingboxException(_takeError(errOut));
      }
    } finally {
      calloc.free(errOut);
    }
  }

  void freeHandle(int handle) {
    final code = _sbFreeHandle(handle);
    if (code != 0) {
      throw SingboxException('invalid handle');
    }
  }

  String _takeString(Pointer<Utf8> pointer) {
    if (pointer == nullptr) {
      return '';
    }
    try {
      return pointer.toDartString();
    } finally {
      _sbFreeString(pointer);
    }
  }

  String _takeError(Pointer<Pointer<Utf8>> errOut) {
    final pointer = errOut.value;
    if (pointer == nullptr) {
      return 'unknown error';
    }
    return _takeString(pointer);
  }
}

class SingboxService {
  SingboxService._(this._ffi, this.handle);

  final SingboxFfi _ffi;
  final int handle;
  bool _closed = false;

  void reload(String configJson) {
    if (_closed) {
      throw SingboxException('service is closed');
    }
    _ffi.reload(handle, configJson);
  }

  void close() {
    if (_closed) {
      return;
    }
    _closed = true;
    try {
      _ffi.stop(handle);
    } finally {
      _ffi.freeHandle(handle);
    }
  }
}
