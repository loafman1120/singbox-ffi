import 'dart:async';
import 'dart:convert';
import 'dart:ffi';
import 'dart:io';

import 'package:ffi/ffi.dart';

/// Native `sb_init_options` struct passed to `sb_init`.
///
/// Most Dart callers should use [SingboxInitOptions] instead. This type is
/// exported for raw FFI users that need the exact C ABI layout.
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

/// Opaque handle returned by `sb_start`.
///
/// A valid handle represents one native sing-box service instance. Close it
/// with [SingboxService.close] or with [SingboxFfi.stop] followed by
/// [SingboxFfi.freeHandle].
typedef SbHandle = int;

typedef SbVersionNative = Pointer<Utf8> Function();
typedef SbVersionDart = Pointer<Utf8> Function();

typedef SbGoVersionNative = Pointer<Utf8> Function();
typedef SbGoVersionDart = Pointer<Utf8> Function();

typedef SbFreeStringNative = Void Function(Pointer<Utf8>);
typedef SbFreeStringDart = void Function(Pointer<Utf8>);

typedef SbInitNative = Int32 Function(
  Pointer<SbInitOptions>,
  Pointer<Pointer<Utf8>>,
);
typedef SbInitDart = int Function(
  Pointer<SbInitOptions>,
  Pointer<Pointer<Utf8>>,
);

typedef SbCheckConfigNative = Int32 Function(
  Pointer<Utf8>,
  Pointer<Pointer<Utf8>>,
);
typedef SbCheckConfigDart = int Function(
  Pointer<Utf8>,
  Pointer<Pointer<Utf8>>,
);

typedef SbStartNative = Int32 Function(
  Pointer<Utf8>,
  Pointer<Uint64>,
  Pointer<Pointer<Utf8>>,
);
typedef SbStartDart = int Function(
  Pointer<Utf8>,
  Pointer<Uint64>,
  Pointer<Pointer<Utf8>>,
);

typedef SbReloadNative = Int32 Function(
  Uint64,
  Pointer<Utf8>,
  Pointer<Pointer<Utf8>>,
);
typedef SbReloadDart = int Function(
  int,
  Pointer<Utf8>,
  Pointer<Pointer<Utf8>>,
);

typedef SbStopNative = Int32 Function(Uint64, Pointer<Pointer<Utf8>>);
typedef SbStopDart = int Function(int, Pointer<Pointer<Utf8>>);

typedef SbFreeHandleNative = Int32 Function(Uint64);
typedef SbFreeHandleDart = int Function(int);

typedef SbDrainLogsNative = Int32 Function(
  Uint64,
  Int32,
  Pointer<Pointer<Utf8>>,
  Pointer<Pointer<Utf8>>,
);
typedef SbDrainLogsDart = int Function(
  int,
  int,
  Pointer<Pointer<Utf8>>,
  Pointer<Pointer<Utf8>>,
);

typedef SbClearLogsNative = Int32 Function(Uint64, Pointer<Pointer<Utf8>>);
typedef SbClearLogsDart = int Function(int, Pointer<Pointer<Utf8>>);

typedef SbServiceStateNative = Int32 Function(
  Uint64,
  Pointer<Pointer<Utf8>>,
  Pointer<Pointer<Utf8>>,
);
typedef SbServiceStateDart = int Function(
  int,
  Pointer<Pointer<Utf8>>,
  Pointer<Pointer<Utf8>>,
);

typedef SbSystemProxyStatusNative = Int32 Function(
  Pointer<Pointer<Utf8>>,
  Pointer<Pointer<Utf8>>,
);
typedef SbSystemProxyStatusDart = int Function(
  Pointer<Pointer<Utf8>>,
  Pointer<Pointer<Utf8>>,
);

typedef SbSetSystemProxyNative = Int32 Function(
  Pointer<Utf8>,
  Int32,
  Pointer<Utf8>,
  Bool,
  Pointer<Pointer<Utf8>>,
);
typedef SbSetSystemProxyDart = int Function(
  Pointer<Utf8>,
  int,
  Pointer<Utf8>,
  bool,
  Pointer<Pointer<Utf8>>,
);

/// Exported native symbol names used by [SingboxRawBindings].
final class SingboxNativeSymbols {
  const SingboxNativeSymbols._();

  static const version = 'sb_version';
  static const goVersion = 'sb_go_version';
  static const freeString = 'sb_free_string';
  static const init = 'sb_init';
  static const checkConfig = 'sb_check_config';
  static const start = 'sb_start';
  static const reload = 'sb_reload';
  static const stop = 'sb_stop';
  static const freeHandle = 'sb_free_handle';
  static const drainLogs = 'sb_drain_logs';
  static const clearLogs = 'sb_clear_logs';
  static const serviceState = 'sb_service_state';
  static const systemProxyStatus = 'sb_system_proxy_status';
  static const setSystemProxy = 'sb_set_system_proxy';
}

/// Typed Dart FFI bindings for the native `singboxffi` C ABI.
///
/// Use this class when you need direct access to the native function pointers.
/// For normal app code, [SingboxFfi] is easier because it converts Dart strings,
/// checks return codes, and frees native strings for you.
class SingboxRawBindings {
  SingboxRawBindings(this.library)
      : sbVersion = library.lookupFunction<SbVersionNative, SbVersionDart>(
          SingboxNativeSymbols.version,
        ),
        sbGoVersion =
            library.lookupFunction<SbGoVersionNative, SbGoVersionDart>(
          SingboxNativeSymbols.goVersion,
        ),
        sbFreeString =
            library.lookupFunction<SbFreeStringNative, SbFreeStringDart>(
          SingboxNativeSymbols.freeString,
        ),
        sbInit = library.lookupFunction<SbInitNative, SbInitDart>(
          SingboxNativeSymbols.init,
        ),
        sbCheckConfig =
            library.lookupFunction<SbCheckConfigNative, SbCheckConfigDart>(
          SingboxNativeSymbols.checkConfig,
        ),
        sbStart = library.lookupFunction<SbStartNative, SbStartDart>(
          SingboxNativeSymbols.start,
        ),
        sbReload = library.lookupFunction<SbReloadNative, SbReloadDart>(
          SingboxNativeSymbols.reload,
        ),
        sbStop = library.lookupFunction<SbStopNative, SbStopDart>(
          SingboxNativeSymbols.stop,
        ),
        sbFreeHandle =
            library.lookupFunction<SbFreeHandleNative, SbFreeHandleDart>(
          SingboxNativeSymbols.freeHandle,
        ),
        sbDrainLogs =
            library.lookupFunction<SbDrainLogsNative, SbDrainLogsDart>(
          SingboxNativeSymbols.drainLogs,
        ),
        sbClearLogs =
            library.lookupFunction<SbClearLogsNative, SbClearLogsDart>(
          SingboxNativeSymbols.clearLogs,
        ),
        sbServiceState =
            library.lookupFunction<SbServiceStateNative, SbServiceStateDart>(
          SingboxNativeSymbols.serviceState,
        ),
        sbSystemProxyStatus = library
            .lookupFunction<SbSystemProxyStatusNative, SbSystemProxyStatusDart>(
          SingboxNativeSymbols.systemProxyStatus,
        ),
        sbSetSystemProxy = library
            .lookupFunction<SbSetSystemProxyNative, SbSetSystemProxyDart>(
          SingboxNativeSymbols.setSystemProxy,
        );

  /// Opens [path], or the platform default library name when [path] is omitted.
  factory SingboxRawBindings.open([String? path]) {
    return SingboxRawBindings(
      DynamicLibrary.open(path ?? SingboxFfi.defaultLibraryName),
    );
  }

  /// Opens a dynamic library using [SingboxFfi.openDefaultLibrary].
  factory SingboxRawBindings.openDefault([String? path]) {
    return SingboxRawBindings(SingboxFfi.openDefaultLibrary(path));
  }

  /// Opens the library using the package's Flutter plugin loading strategy.
  factory SingboxRawBindings.openBundled([String? path]) {
    return SingboxRawBindings(SingboxFfi.openBundledLibrary(path));
  }

  /// Resolves symbols from the current process.
  ///
  /// This only works when the native archive or framework has already been
  /// linked into the app process.
  factory SingboxRawBindings.process() {
    return SingboxRawBindings(DynamicLibrary.process());
  }

  final DynamicLibrary library;
  final SbVersionDart sbVersion;
  final SbGoVersionDart sbGoVersion;
  final SbFreeStringDart sbFreeString;
  final SbInitDart sbInit;
  final SbCheckConfigDart sbCheckConfig;
  final SbStartDart sbStart;
  final SbReloadDart sbReload;
  final SbStopDart sbStop;
  final SbFreeHandleDart sbFreeHandle;
  final SbDrainLogsDart sbDrainLogs;
  final SbClearLogsDart sbClearLogs;
  final SbServiceStateDart sbServiceState;
  final SbSystemProxyStatusDart sbSystemProxyStatus;
  final SbSetSystemProxyDart sbSetSystemProxy;
}

/// Broad category for errors reported by the high-level wrapper.
enum SingboxErrorKind {
  config,
  permission,
  missingDependency,
  systemProxyUnavailable,
  serviceState,
  native,
  unknown,
}

/// Error thrown by the high-level Dart wrapper.
class SingboxException implements Exception {
  SingboxException(
    this.message, {
    this.kind = SingboxErrorKind.unknown,
    this.code,
  });

  final String message;
  final SingboxErrorKind kind;
  final String? code;

  @override
  String toString() {
    final codeSuffix = code == null ? '' : ', code: $code';
    return 'SingboxException(${kind.name}$codeSuffix): $message';
  }
}

/// Initialization options for [SingboxFfi.init].
///
/// These values are forwarded to libbox setup before any service starts. The
/// defaults are intentionally small and local so command-line examples can run
/// without app-specific directories.
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

  /// Base directory used by libbox.
  final String basePath;

  /// Directory used for runtime state.
  final String workingPath;

  /// Directory used for temporary runtime files.
  final String tempPath;

  /// Optional locale passed to libbox before setup.
  final String? locale;

  /// Secret used by libbox command server support.
  final String commandSecret;

  /// Command server listen port. Use `0` for automatic/default behavior.
  final int commandPort;

  /// Maximum number of log lines retained by libbox internals.
  final int logMaxLines;

  /// Enables libbox debug setup.
  final bool debug;

  /// Enables libbox OOM killer behavior.
  final bool oomKillerEnabled;

  /// Disables libbox OOM killer behavior.
  final bool oomKillerDisabled;

  /// Optional OOM memory limit passed through to libbox.
  final int oomMemoryLimit;
}

/// One log event emitted by a running sing-box service.
///
/// A reset event means libbox cleared its internal log history, usually because
/// a service started or reloaded. UI log panels can use [isReset] to clear the
/// visible list before appending later entries.
class SingboxLogEvent {
  const SingboxLogEvent({
    required this.sequence,
    required this.level,
    required this.levelName,
    required this.message,
    required this.isReset,
  });

  factory SingboxLogEvent.fromJson(Map<String, Object?> json) {
    return SingboxLogEvent(
      sequence: (json['sequence'] as num?)?.toInt() ?? 0,
      level: (json['level'] as num?)?.toInt() ?? 0,
      levelName: json['levelName'] as String? ?? '',
      message: json['message'] as String? ?? '',
      isReset: json['reset'] == true,
    );
  }

  /// Monotonic native-side sequence number for this handle.
  final int sequence;

  /// Numeric sing-box log level.
  final int level;

  /// Human-readable level name such as `info`, `warn`, or `error`.
  final String levelName;

  /// Formatted log message from sing-box/libbox.
  final String message;

  /// Whether this event means the native log history was reset.
  final bool isReset;

  /// Whether this event contains a normal log message.
  bool get isMessage => !isReset;

  @override
  String toString() {
    if (isReset) {
      return 'SingboxLogEvent.reset(sequence: $sequence)';
    }
    return 'SingboxLogEvent($levelName, $message)';
  }
}

/// Native service runtime state.
enum SingboxServiceState {
  running,
  stopped,
  closed,
  unknown;

  static SingboxServiceState fromJson(String? value) {
    return switch (value) {
      'running' => SingboxServiceState.running,
      'stopped' => SingboxServiceState.stopped,
      'closed' => SingboxServiceState.closed,
      _ => SingboxServiceState.unknown,
    };
  }
}

/// Snapshot of one native service handle.
class SingboxServiceSnapshot {
  const SingboxServiceSnapshot({
    required this.state,
    required this.running,
    required this.closed,
    this.lastError,
  });

  factory SingboxServiceSnapshot.fromJson(Map<String, Object?> json) {
    return SingboxServiceSnapshot(
      state: SingboxServiceState.fromJson(json['state'] as String?),
      running: json['running'] == true,
      closed: json['closed'] == true,
      lastError: json['lastError'] as String?,
    );
  }

  final SingboxServiceState state;
  final bool running;
  final bool closed;
  final String? lastError;

  @override
  String toString() {
    return 'SingboxServiceSnapshot(state: ${state.name}, running: $running, '
        'closed: $closed, lastError: $lastError)';
  }
}

/// System proxy state reported by the native platform layer.
class SingboxSystemProxyStatus {
  const SingboxSystemProxyStatus({
    required this.platform,
    required this.supported,
    required this.enabled,
    this.server,
    this.bypass,
    required this.hasSavedState,
  });

  factory SingboxSystemProxyStatus.fromJson(Map<String, Object?> json) {
    return SingboxSystemProxyStatus(
      platform: json['platform'] as String? ?? '',
      supported: json['supported'] == true,
      enabled: json['enabled'] == true,
      server: json['server'] as String?,
      bypass: json['bypass'] as String?,
      hasSavedState: json['hasSavedState'] == true,
    );
  }

  final String platform;
  final bool supported;
  final bool enabled;
  final String? server;
  final String? bypass;
  final bool hasSavedState;

  @override
  String toString() {
    return 'SingboxSystemProxyStatus(platform: $platform, '
        'supported: $supported, enabled: $enabled, server: $server)';
  }
}

/// Target endpoint used when enabling the OS system proxy.
class SingboxSystemProxyOptions {
  const SingboxSystemProxyOptions({
    this.host = '127.0.0.1',
    this.port = 2080,
    this.bypass = '<local>',
  });

  final String host;
  final int port;
  final String bypass;
}

/// High-level Dart wrapper around the native `singboxffi` library.
///
/// A typical lifecycle is:
///
/// 1. Open the library with [SingboxFfi.openBundled].
/// 2. Call [init] once for the process.
/// 3. Validate JSON with [checkConfig].
/// 4. Start a service with [start].
/// 5. Reload or close the returned [SingboxService].
class SingboxFfi {
  SingboxFfi._(this.raw);

  /// Creates a wrapper from an already-opened dynamic library.
  factory SingboxFfi.fromLibrary(DynamicLibrary library) {
    return SingboxFfi._(SingboxRawBindings(library));
  }

  /// Opens [path], or the platform default library name when [path] is omitted.
  factory SingboxFfi.open([String? path]) {
    return SingboxFfi._(SingboxRawBindings.open(path));
  }

  /// Opens a dynamic library from common default search locations.
  factory SingboxFfi.openDefault([String? path]) {
    return SingboxFfi._(SingboxRawBindings.openDefault(path));
  }

  /// Opens the library using the package's Flutter plugin loading strategy.
  ///
  /// This is the recommended constructor for Flutter apps. It uses [path] when
  /// provided, resolves process symbols on iOS, and otherwise falls back to the
  /// platform dynamic library search used by [openDefault].
  factory SingboxFfi.openBundled([String? path]) {
    return SingboxFfi._(SingboxRawBindings.openBundled(path));
  }

  /// Resolves native symbols from the current process.
  ///
  /// Use this only when the native library has already been linked into the app
  /// executable, runner, or platform plugin.
  factory SingboxFfi.process() {
    return SingboxFfi._(SingboxRawBindings.process());
  }

  /// Platform default dynamic library file name.
  static String get defaultLibraryName {
    if (Platform.isWindows) {
      return 'singboxffi.dll';
    }
    if (Platform.isMacOS || Platform.isIOS) {
      return 'libsingboxffi.dylib';
    }
    return 'libsingboxffi.so';
  }

  /// Opens [path] or searches common dynamic library locations.
  ///
  /// Search order without [path]:
  ///
  /// 1. Current working directory.
  /// 2. `Platform.executable` directory.
  /// 3. `Platform.resolvedExecutable` directory.
  /// 4. Platform loader default using [defaultLibraryName].
  static DynamicLibrary openDefaultLibrary([String? path]) {
    if (path != null && path.isNotEmpty) {
      return DynamicLibrary.open(path);
    }

    final libraryName = defaultLibraryName;
    final searched = <String>{};
    final candidates = <String>[
      _joinPath(Directory.current.path, libraryName),
      _joinPath(File(Platform.executable).parent.path, libraryName),
      _joinPath(File(Platform.resolvedExecutable).parent.path, libraryName),
    ];

    Object? lastError;
    for (final candidate in candidates) {
      if (!searched.add(candidate)) {
        continue;
      }
      try {
        return DynamicLibrary.open(candidate);
      } catch (error) {
        lastError = error;
      }
    }

    try {
      return DynamicLibrary.open(libraryName);
    } catch (error) {
      throw SingboxException(
        'failed to open $libraryName from default search paths: '
        '${lastError ?? error}',
        kind: SingboxErrorKind.missingDependency,
        code: 'native.library_not_found',
      );
    }
  }

  /// Opens the library using the Flutter plugin's bundled artifact strategy.
  static DynamicLibrary openBundledLibrary([String? path]) {
    if (path != null && path.isNotEmpty) {
      return DynamicLibrary.open(path);
    }
    if (Platform.isIOS) {
      return DynamicLibrary.process();
    }
    return openDefaultLibrary();
  }

  static String _joinPath(String directory, String fileName) {
    if (directory.isEmpty || directory == '.') {
      return fileName;
    }
    final separator = Platform.pathSeparator;
    if (directory.endsWith(separator) ||
        directory.endsWith('/') ||
        directory.endsWith('\\')) {
      return '$directory$fileName';
    }
    return '$directory$separator$fileName';
  }

  /// Raw native bindings used by this wrapper.
  final SingboxRawBindings raw;

  /// Returns the sing-box/libbox version reported by the native core.
  String version() => takeString(raw.sbVersion());

  /// Returns the Go runtime version reported by the native core.
  String goVersion() => takeString(raw.sbGoVersion());

  /// Initializes libbox with [options].
  ///
  /// Call this before [checkConfig] or [start]. Native errors are converted to
  /// [SingboxException].
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

      final code = raw.sbInit(opts, errOut);
      if (code != 0) {
        throw takeNativeException(errOut, fallbackCode: 'native.init_failed');
      }
    } finally {
      for (final pointer in allocations) {
        calloc.free(pointer);
      }
      calloc.free(errOut);
      calloc.free(opts);
    }
  }

  /// Validates a sing-box JSON configuration.
  ///
  /// Throws [SingboxException] when libbox rejects the config.
  void checkConfig(String configJson) {
    final config = configJson.toNativeUtf8(allocator: calloc);
    final errOut = calloc<Pointer<Utf8>>();
    try {
      final code = raw.sbCheckConfig(config, errOut);
      if (code != 0) {
        throw takeNativeException(
          errOut,
          fallbackKind: SingboxErrorKind.config,
          fallbackCode: 'config.invalid',
        );
      }
    } finally {
      calloc.free(config);
      calloc.free(errOut);
    }
  }

  /// Starts a sing-box service from [configJson].
  ///
  /// The returned [SingboxService] owns the native handle and should be closed
  /// when the service is no longer needed.
  SingboxService start(
    String configJson, {
    bool systemProxy = false,
    SingboxSystemProxyOptions systemProxyOptions =
        const SingboxSystemProxyOptions(),
  }) {
    final config = configJson.toNativeUtf8(allocator: calloc);
    final handleOut = calloc<Uint64>();
    final errOut = calloc<Pointer<Utf8>>();
    try {
      final code = raw.sbStart(config, handleOut, errOut);
      if (code != 0) {
        throw takeNativeException(errOut, fallbackCode: 'service.start_failed');
      }
      final service = SingboxService._(this, handleOut.value);
      if (systemProxy) {
        try {
          service.enableSystemProxy(systemProxyOptions);
        } catch (_) {
          service.close();
          rethrow;
        }
      }
      return service;
    } finally {
      calloc.free(config);
      calloc.free(handleOut);
      calloc.free(errOut);
    }
  }

  /// Reloads an existing native service [handle] with [configJson].
  void reload(SbHandle handle, String configJson) {
    final config = configJson.toNativeUtf8(allocator: calloc);
    final errOut = calloc<Pointer<Utf8>>();
    try {
      final code = raw.sbReload(handle, config, errOut);
      if (code != 0) {
        throw takeNativeException(
          errOut,
          fallbackCode: 'service.reload_failed',
        );
      }
    } finally {
      calloc.free(config);
      calloc.free(errOut);
    }
  }

  /// Stops an existing native service [handle].
  ///
  /// This does not remove the handle from the native handle table; call
  /// [freeHandle] after stopping, or use [SingboxService.close].
  void stop(SbHandle handle) {
    final errOut = calloc<Pointer<Utf8>>();
    try {
      final code = raw.sbStop(handle, errOut);
      if (code != 0) {
        throw takeNativeException(errOut, fallbackCode: 'service.stop_failed');
      }
    } finally {
      calloc.free(errOut);
    }
  }

  /// Releases a stopped native service [handle].
  void freeHandle(SbHandle handle) {
    final code = raw.sbFreeHandle(handle);
    if (code != 0) {
      throw SingboxException(
        'invalid handle',
        kind: SingboxErrorKind.serviceState,
        code: 'service.invalid_handle',
      );
    }
  }

  /// Drains queued log events for [handle].
  ///
  /// Set [maxEntries] to a positive value to limit the number of events drained
  /// in one call. The default `0` drains everything currently queued.
  List<SingboxLogEvent> drainLogs(SbHandle handle, {int maxEntries = 0}) {
    final jsonOut = calloc<Pointer<Utf8>>();
    final errOut = calloc<Pointer<Utf8>>();
    try {
      final code = raw.sbDrainLogs(handle, maxEntries, jsonOut, errOut);
      if (code != 0) {
        throw takeNativeException(
          errOut,
          fallbackCode: 'service.drain_logs_failed',
        );
      }
      final payload = takeString(jsonOut.value);
      final decoded = jsonDecode(payload) as List<Object?>;
      return decoded
          .map((item) => SingboxLogEvent.fromJson(
                Map<String, Object?>.from(item as Map),
              ))
          .toList(growable: false);
    } finally {
      calloc.free(jsonOut);
      calloc.free(errOut);
    }
  }

  /// Clears queued Dart-facing logs and libbox's internal log history.
  void clearLogs(SbHandle handle) {
    final errOut = calloc<Pointer<Utf8>>();
    try {
      final code = raw.sbClearLogs(handle, errOut);
      if (code != 0) {
        throw takeNativeException(
          errOut,
          fallbackCode: 'service.clear_logs_failed',
        );
      }
    } finally {
      calloc.free(errOut);
    }
  }

  /// Returns a native state snapshot for [handle].
  SingboxServiceSnapshot serviceState(SbHandle handle) {
    final jsonOut = calloc<Pointer<Utf8>>();
    final errOut = calloc<Pointer<Utf8>>();
    try {
      final code = raw.sbServiceState(handle, jsonOut, errOut);
      if (code != 0) {
        throw takeNativeException(
          errOut,
          fallbackCode: 'service.state_failed',
        );
      }
      final payload = takeString(jsonOut.value);
      return SingboxServiceSnapshot.fromJson(
        Map<String, Object?>.from(jsonDecode(payload) as Map),
      );
    } finally {
      calloc.free(jsonOut);
      calloc.free(errOut);
    }
  }

  /// Returns the current OS system proxy state.
  SingboxSystemProxyStatus systemProxyStatus() {
    final jsonOut = calloc<Pointer<Utf8>>();
    final errOut = calloc<Pointer<Utf8>>();
    try {
      final code = raw.sbSystemProxyStatus(jsonOut, errOut);
      if (code != 0) {
        throw takeNativeException(
          errOut,
          fallbackKind: SingboxErrorKind.systemProxyUnavailable,
          fallbackCode: 'system_proxy.status_failed',
        );
      }
      final payload = takeString(jsonOut.value);
      return SingboxSystemProxyStatus.fromJson(
        Map<String, Object?>.from(jsonDecode(payload) as Map),
      );
    } finally {
      calloc.free(jsonOut);
      calloc.free(errOut);
    }
  }

  /// Enables or restores the OS system proxy.
  void setSystemProxy(
    bool enabled, {
    SingboxSystemProxyOptions options = const SingboxSystemProxyOptions(),
  }) {
    final host = options.host.toNativeUtf8(allocator: calloc);
    final bypass = options.bypass.toNativeUtf8(allocator: calloc);
    final errOut = calloc<Pointer<Utf8>>();
    try {
      final code = raw.sbSetSystemProxy(
        host,
        options.port,
        bypass,
        enabled,
        errOut,
      );
      if (code != 0) {
        throw takeNativeException(
          errOut,
          fallbackKind: SingboxErrorKind.systemProxyUnavailable,
          fallbackCode: enabled
              ? 'system_proxy.enable_failed'
              : 'system_proxy.restore_failed',
        );
      }
    } finally {
      calloc.free(host);
      calloc.free(bypass);
      calloc.free(errOut);
    }
  }

  /// Polls native logs for [handle] and exposes them as a Dart stream.
  ///
  /// The stream drains pending native events every [interval]. It is broadcast
  /// so multiple UI widgets can listen to the same stream returned by app code.
  Stream<SingboxLogEvent> logs(
    SbHandle handle, {
    Duration interval = const Duration(milliseconds: 250),
    int batchSize = 0,
  }) {
    late StreamController<SingboxLogEvent> controller;
    Timer? timer;

    void drain() {
      for (final event in drainLogs(handle, maxEntries: batchSize)) {
        controller.add(event);
      }
    }

    controller = StreamController<SingboxLogEvent>.broadcast(
      onListen: () {
        drain();
        timer ??= Timer.periodic(interval, (_) => drain());
      },
      onCancel: () {
        if (!controller.hasListener) {
          timer?.cancel();
          timer = null;
        }
      },
    );
    return controller.stream;
  }

  /// Frees a native string returned by the core.
  void freeString(Pointer<Utf8> pointer) {
    if (pointer != nullptr) {
      raw.sbFreeString(pointer);
    }
  }

  /// Converts an owned native string to Dart and frees the native pointer.
  String takeString(Pointer<Utf8> pointer) {
    if (pointer == nullptr) {
      return '';
    }
    try {
      return pointer.toDartString();
    } finally {
      freeString(pointer);
    }
  }

  /// Takes an error string from an `err_out` pointer and frees it.
  String takeError(Pointer<Pointer<Utf8>> errOut) {
    final pointer = errOut.value;
    if (pointer == nullptr) {
      return 'unknown error';
    }
    return takeString(pointer);
  }

  /// Converts a native `err_out` pointer to a structured [SingboxException].
  SingboxException takeNativeException(
    Pointer<Pointer<Utf8>> errOut, {
    SingboxErrorKind fallbackKind = SingboxErrorKind.native,
    String? fallbackCode,
  }) {
    final message = takeError(errOut);
    final classified = _classifyNativeError(
      message,
      fallbackKind: fallbackKind,
      fallbackCode: fallbackCode,
    );
    return SingboxException(
      message,
      kind: classified.kind,
      code: classified.code,
    );
  }

  static ({SingboxErrorKind kind, String? code}) _classifyNativeError(
    String message, {
    required SingboxErrorKind fallbackKind,
    String? fallbackCode,
  }) {
    final lower = message.toLowerCase();
    if (lower.contains('invalid handle') ||
        lower.contains('service is closed') ||
        lower.contains('os: invalid') ||
        lower.contains('file already closed')) {
      return (
        kind: SingboxErrorKind.serviceState,
        code: 'service.invalid_state',
      );
    }
    if (lower.contains('permission') ||
        lower.contains('access is denied') ||
        lower.contains('administrator') ||
        lower.contains('privilege')) {
      return (
        kind: SingboxErrorKind.permission,
        code: 'platform.permission_denied',
      );
    }
    if (lower.contains('system proxy')) {
      return (
        kind: SingboxErrorKind.systemProxyUnavailable,
        code: 'system_proxy.unavailable',
      );
    }
    if (lower.contains('not found') ||
        lower.contains('no such file') ||
        lower.contains('cannot find')) {
      return (
        kind: SingboxErrorKind.missingDependency,
        code: 'native.missing_dependency',
      );
    }
    return (kind: fallbackKind, code: fallbackCode);
  }
}

/// Running sing-box service handle owned by [SingboxFfi].
///
/// Call [close] exactly once when done. Repeated [close] calls are ignored.
class SingboxService {
  SingboxService._(this._ffi, this.handle);

  final SingboxFfi _ffi;

  /// Native service handle.
  final SbHandle handle;
  bool _closed = false;
  bool _systemProxyEnabled = false;

  /// Whether this Dart wrapper has already released its native handle.
  bool get isClosed => _closed;

  /// Returns the current native state snapshot for this service.
  SingboxServiceSnapshot state() {
    if (_closed) {
      return const SingboxServiceSnapshot(
        state: SingboxServiceState.closed,
        running: false,
        closed: true,
      );
    }
    return _ffi.serviceState(handle);
  }

  /// Reloads this service with a new sing-box JSON configuration.
  void reload(String configJson) {
    if (_closed) {
      throw SingboxException(
        'service is closed',
        kind: SingboxErrorKind.serviceState,
        code: 'service.closed',
      );
    }
    _ffi.reload(handle, configJson);
  }

  /// Drains queued log events for this service.
  List<SingboxLogEvent> drainLogs({int maxEntries = 0}) {
    if (_closed) {
      throw SingboxException(
        'service is closed',
        kind: SingboxErrorKind.serviceState,
        code: 'service.closed',
      );
    }
    return _ffi.drainLogs(handle, maxEntries: maxEntries);
  }

  /// Clears queued logs and libbox's internal log history for this service.
  void clearLogs() {
    if (_closed) {
      throw SingboxException(
        'service is closed',
        kind: SingboxErrorKind.serviceState,
        code: 'service.closed',
      );
    }
    _ffi.clearLogs(handle);
  }

  /// Streams log events from this service.
  Stream<SingboxLogEvent> logs({
    Duration interval = const Duration(milliseconds: 250),
    int batchSize = 0,
  }) {
    if (_closed) {
      throw SingboxException(
        'service is closed',
        kind: SingboxErrorKind.serviceState,
        code: 'service.closed',
      );
    }
    return _ffi.logs(handle, interval: interval, batchSize: batchSize);
  }

  /// Enables the OS system proxy for this running service.
  ///
  /// The proxy is restored automatically when [close] is called.
  void enableSystemProxy([
    SingboxSystemProxyOptions options = const SingboxSystemProxyOptions(),
  ]) {
    if (_closed) {
      throw SingboxException(
        'service is closed',
        kind: SingboxErrorKind.serviceState,
        code: 'service.closed',
      );
    }
    _ffi.setSystemProxy(true, options: options);
    _systemProxyEnabled = true;
  }

  /// Restores the OS system proxy if this service enabled it.
  void disableSystemProxy() {
    if (!_systemProxyEnabled) {
      return;
    }
    _ffi.setSystemProxy(false);
    _systemProxyEnabled = false;
  }

  /// Stops the native service and releases its handle.
  void close() {
    if (_closed) {
      return;
    }
    _closed = true;
    try {
      if (_systemProxyEnabled) {
        try {
          _ffi.setSystemProxy(false);
          _systemProxyEnabled = false;
        } catch (_) {
          // Still stop and free the native service; callers can query/restore
          // system proxy explicitly if platform restoration failed.
        }
      }
      _ffi.stop(handle);
    } finally {
      _ffi.freeHandle(handle);
    }
  }
}
