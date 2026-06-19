cmake_minimum_required(VERSION 3.14)
project(singbox_ffi_plugin LANGUAGES C)

string(TOLOWER "${CMAKE_SYSTEM_PROCESSOR}" SINGBOX_FFI_PROCESSOR)

if(SINGBOX_FFI_PROCESSOR MATCHES "arm64|aarch64")
  set(SINGBOX_FFI_ARCH "arm64")
elseif(CMAKE_SIZEOF_VOID_P EQUAL 8)
  set(SINGBOX_FFI_ARCH "x64")
else()
  set(SINGBOX_FFI_ARCH "x86")
endif()

set(SINGBOX_FFI_ARTIFACT_DIR "${CMAKE_CURRENT_SOURCE_DIR}/artifacts")
set(SINGBOX_FFI_PREBUILT "${SINGBOX_FFI_ARTIFACT_DIR}/${SINGBOX_FFI_ARCH}/singboxffi.dll")

if(NOT EXISTS "${SINGBOX_FFI_PREBUILT}")
  message(FATAL_ERROR
    "Missing ${SINGBOX_FFI_PREBUILT}. "
    "Place the prebuilt singboxffi.dll in windows/artifacts/${SINGBOX_FFI_ARCH}/.")
endif()

set(singbox_ffi_bundled_libraries
  "${SINGBOX_FFI_PREBUILT}"
  PARENT_SCOPE
)
