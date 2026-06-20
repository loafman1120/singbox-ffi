Pod::Spec.new do |s|
  force_load_archive = File.exist?(File.join(__dir__, 'Libraries', 'libsingboxffi.a'))

  s.name             = 'singbox_ffi'
  s.version          = '0.1.2'
  s.summary          = 'Flutter FFI packaging for the singbox-ffi native core.'
  s.description      = 'Packages and links the singbox-ffi native core for Flutter apps.'
  s.homepage         = 'https://github.com/loafman1120/singbox-ffi'
  s.license          = { :file => '../LICENSE' }
  s.author           = { 'loafman1120' => 'loafman1120@users.noreply.github.com' }
  s.source           = { :path => '.' }
  s.source_files     = 'Classes/**/*'
  s.public_header_files = 'Classes/**/*.h'
  s.vendored_libraries = ['Libraries/**/*.a']
  s.vendored_frameworks = ['Frameworks/**/*.framework', 'Frameworks/**/*.xcframework']
  s.platform = :ios, '12.0'
  s.pod_target_xcconfig = {
    'DEFINES_MODULE' => 'YES',
    'OTHER_LDFLAGS' => '$(inherited) -lc++' + (force_load_archive ? ' -force_load "$(PODS_TARGET_SRCROOT)/Libraries/libsingboxffi.a"' : '')
  }
end
