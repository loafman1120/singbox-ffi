Pod::Spec.new do |s|
  s.name             = 'singbox_ffi'
  s.version          = '0.1.0'
  s.summary          = 'Flutter FFI packaging for the singbox-ffi native core.'
  s.description      = 'Packages and links the singbox-ffi native core for Flutter macOS apps.'
  s.homepage         = 'https://github.com/loafman1120/singbox-ffi'
  s.license          = { :file => '../LICENSE' }
  s.author           = { 'loafman1120' => 'loafman1120@users.noreply.github.com' }
  s.source           = { :path => '.' }
  s.source_files     = 'Classes/**/*'
  s.public_header_files = 'Classes/**/*.h'
  s.vendored_libraries = ['Libraries/**/*.a', 'Libraries/**/*.dylib']
  s.vendored_frameworks = ['Frameworks/**/*.framework']
  s.platform = :osx, '10.14'
  s.pod_target_xcconfig = {
    'DEFINES_MODULE' => 'YES',
    'OTHER_LDFLAGS' => '$(inherited) -lc++'
  }
end
