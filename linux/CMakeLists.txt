cmake_minimum_required(VERSION 3.10)
project(singbox_ffi_plugin LANGUAGES C)

set(SINGBOX_FFI_ARTIFACT_DIR "${CMAKE_CURRENT_SOURCE_DIR}/artifacts")
set(SINGBOX_FFI_PREBUILT "${SINGBOX_FFI_ARTIFACT_DIR}/${CMAKE_SYSTEM_PROCESSOR}/libsingboxffi.so")

if(NOT EXISTS "${SINGBOX_FFI_PREBUILT}")
  message(FATAL_ERROR
    "Missing ${SINGBOX_FFI_PREBUILT}. "
    "Place the prebuilt libsingboxffi.so in linux/artifacts/${CMAKE_SYSTEM_PROCESSOR}/.")
endif()

set(singbox_ffi_bundled_libraries
  "${SINGBOX_FFI_PREBUILT}"
  PARENT_SCOPE
)
