var os = require('os');
var loadGruntTasks = require('load-grunt-tasks');
const webpackDevConfig = require('./webpack/webpack.develop');
const webpackProdConfig = require('./webpack/webpack.production');

var arch = os.arch();
if (arch === 'x64') arch = 'amd64';

var portainer_data = '/tmp/portainer';

module.exports = function (grunt) {
  loadGruntTasks(grunt, {
    pattern: ['grunt-*', 'gruntify-*'],
  });

  grunt.initConfig({
    root: 'dist',
    distdir: 'dist/public',
    binaries: {
      dockerLinuxVersion: '19.03.13',
      dockerWindowsVersion: '19-03-12',
      dockerLinuxComposeVersion: '1.28.2',
      dockerWindowsComposeVersion: '1.28.0',
      komposeVersion: 'v1.22.0',
      kubectlVersion: 'v1.18.0',
    },
    config: gruntfile_cfg.config,
    env: gruntfile_cfg.env,
    src: gruntfile_cfg.src,
    clean: gruntfile_cfg.clean,
    eslint: gruntfile_cfg.eslint,
    shell: gruntfile_cfg.shell,
    copy: gruntfile_cfg.copy,
    webpack: gruntfile_cfg.webpack,
  });

  grunt.registerTask('lint', ['eslint']);

  grunt.registerTask('build:server', [
    'shell:build_binary:linux:' + arch,
    'shell:download_docker_binary:linux:' + arch,
    'shell:download_docker_compose_binary:linux:' + arch,
    'shell:download_kompose_binary:linux:' + arch,
    'shell:download_kubectl_binary:linux:' + arch,
  ]);

  grunt.registerTask('build:client', ['config:dev', 'env:dev', 'webpack:dev']);

  grunt.registerTask('build', ['build:server', 'build:client', 'copy:assets']);

  grunt.registerTask('start:server', ['build:server', 'copy:assets', 'shell:run_container']);

  grunt.registerTask('start:localserver', ['shell:build_binary:linux:' + arch, 'shell:run_localserver']);

  grunt.registerTask('start:client', ['shell:install_yarndeps', 'config:dev', 'env:dev', 'webpack:devWatch']);

  grunt.registerTask('start', ['start:server', 'start:client']);

  grunt.registerTask('start:toolkit', ['start:localserver', 'start:client']);

  grunt.task.registerTask('release', 'release:<platform>:<arch>', function (p = 'linux', a = arch) {
    grunt.task.run([
      'config:prod',
      'env:prod',
      'clean:all',
      'copy:assets',
      'shell:build_binary:' + p + ':' + a,
      'shell:download_docker_binary:' + p + ':' + a,
      'shell:download_docker_compose_binary:' + p + ':' + a,
      'shell:download_kompose_binary:' + p + ':' + a,
      'shell:download_kubectl_binary:' + p + ':' + a,
      'webpack:prod',
    ]);
  });

  grunt.task.registerTask('devopsbuild', 'devopsbuild:<platform>:<arch>', function (p, a) {
    grunt.task.run([
      'config:prod',
      'env:prod',
      'clean:all',
      'copy:assets',
      'shell:build_binary_azuredevops:' + p + ':' + a,
      'shell:download_docker_binary:' + p + ':' + a,
      'shell:download_docker_compose_binary:' + p + ':' + a,
      'shell:download_kompose_binary:' + p + ':' + a,
      'shell:download_kubectl_binary:' + p + ':' + a,
      'webpack:prod',
    ]);
  });
};

/***/
var gruntfile_cfg = {};

gruntfile_cfg.env = {
  dev: {
    NODE_ENV: 'development',
  },
  prod: {
    NODE_ENV: 'production',
  },
};

gruntfile_cfg.webpack = {
  dev: webpackDevConfig,
  devWatch: Object.assign({ watch: true }, webpackDevConfig),
  prod: webpackProdConfig,
};

gruntfile_cfg.config = {
  dev: { options: { variables: { environment: 'development' } } },
  prod: { options: { variables: { environment: 'production' } } },
};

gruntfile_cfg.src = {
  js: ['app/**/__module.js', 'app/**/*.js', '!app/**/*.spec.js'],
  jsTpl: ['<%= distdir %>/templates/**/*.js'],
  html: ['index.html'],
  tpl: ['app/**/*.html'],
  css: ['assets/css/app.css', 'app/**/*.css'],
};

gruntfile_cfg.clean = {
  server: ['<%= root %>/portainer'],
  client: ['<%= distdir %>/*'],
  all: ['<%= root %>/*'],
};

gruntfile_cfg.eslint = {
  src: ['gruntfile.js', '<%= src.js %>'],
  options: { configFile: '.eslintrc.yml' },
};

gruntfile_cfg.copy = {
  assets: {
    files: [],
  },
};

gruntfile_cfg.shell = {
  build_binary: { command: shell_build_binary },
  build_binary_azuredevops: { command: shell_build_binary_azuredevops },
  download_docker_binary: { command: shell_download_docker_binary },
  download_kompose_binary: { command: shell_download_kompose_binary },
  download_kubectl_binary: { command: shell_download_kubectl_binary },
  download_docker_compose_binary: { command: shell_download_docker_compose_binary },
  run_container: { command: shell_run_container },
  run_localserver: { command: shell_run_localserver, options: { async: true } },
  install_yarndeps: { command: shell_install_yarndeps },
};

function shell_build_binary(p, a) {
  var binfile = 'dist/portainer';
  if (p === 'linux') {
    return ['if [ -f ' + binfile + ' ]; then', 'echo "Portainer binary exists";', 'else', 'build/build_binary.sh ' + p + ' ' + a + ';', 'fi'].join(' ');
  } else {
    return [
      'powershell -Command "& {if (Get-Item -Path ' + binfile + '.exe -ErrorAction:SilentlyContinue) {',
      'Write-Host "Portainer binary exists"',
      '} else {',
      '& ".\\build\\build_binary.ps1" -platform ' + p + ' -arch ' + a + '',
      '}}"',
    ].join(' ');
  }
}

function shell_build_binary_azuredevops(p, a) {
  if (p === 'linux') {
    return 'build/build_binary_azuredevops.sh ' + p + ' ' + a + ';';
  } else {
    return 'powershell -Command ".\\build\\build_binary_azuredevops.ps1 -platform ' + p + ' -arch ' + a + '"';
  }
}

function shell_run_container() {
  return [
    'docker rm -f portainer',
    'docker run -d -p 8000:8000 -p 9000:9000 -v $(pwd)/dist:/app -v ' +
      portainer_data +
      ':/data -v /var/run/docker.sock:/var/run/docker.sock:z -v /tmp:/tmp --name portainer portainer/base /app/portainer',
  ].join(';');
}

function shell_run_localserver() {
  return './dist/portainer';
}

function shell_install_yarndeps() {
  return 'yarn';
}

function shell_download_docker_binary(p, a) {
  var ps = { windows: 'win', darwin: 'mac' };
  var as = { amd64: 'x86_64', arm: 'armhf', arm64: 'aarch64' };
  var ip = ps[p] === undefined ? p : ps[p];
  var ia = as[a] === undefined ? a : as[a];
  var binaryVersion = p === 'windows' ? '<%= binaries.dockerWindowsVersion %>' : '<%= binaries.dockerLinuxVersion %>';

  if (p === 'linux' || p === 'mac') {
    return ['if [ -f dist/docker ]; then', 'echo "docker binary exists";', 'else', 'build/download_docker_binary.sh ' + ip + ' ' + ia + ' ' + binaryVersion + ';', 'fi'].join(' ');
  } else {
    return [
      'powershell -Command "& {if (Test-Path -Path "dist/docker.exe") {',
      'Write-Host "Skipping download, Docker binary exists"',
      'return',
      '} else {',
      '& ".\\build\\download_docker_binary.ps1" -docker_version ' + binaryVersion + '',
      '}}"',
    ].join(' ');
  }
}

function shell_download_docker_compose_binary(p, a) {
  var ps = { windows: 'win', darwin: 'mac' };
  var as = { arm: 'armhf', arm64: 'aarch64' };
  var ip = ps[p] || p;
  var ia = as[a] || a;
  var linuxBinaryVersion = '<%= binaries.dockerLinuxComposeVersion %>';
  var windowsBinaryVersion = '<%= binaries.dockerWindowsComposeVersion %>';

  if (ip === 'linux' || ip === 'mac') {
    return [
      'if [ -f dist/docker-compose ]; then',
      'echo "Docker Compose binary exists";',
      'else',
      'build/download_docker_compose_binary.sh ' + ip + ' ' + ia + ' ' + linuxBinaryVersion + ';',
      'fi',
    ].join(' ');
  } else if (ip === 'win') {
    return [
      'powershell -Command "& {if (Test-Path -Path "dist/docker-compose.exe") {',
      'Write-Host "Skipping download, Docker Compose binary exists"',
      'return',
      '} else {',
      '& ".\\build\\download_docker_compose_binary.ps1" -docker_compose_version ' + windowsBinaryVersion + '',
      '}}"',
    ].join(' ');
  }
}

function shell_download_kompose_binary(p, a) {
  var binaryVersion = '<%= binaries.komposeVersion %>';

  if (p === 'linux' || p === 'darwin') {
    return ['if [ -f dist/kompose ]; then', 'echo "kompose binary exists";', 'else', 'build/download_kompose_binary.sh ' + p + ' ' + a + ' ' + binaryVersion + ';', 'fi'].join(' ');
  } else {
    return [
      'powershell -Command "& {if (Test-Path -Path "dist/kompose.exe") {',
      'Write-Host "Skipping download, Kompose binary exists"',
      'return',
      '} else {',
      '& ".\\build\\download_kompose_binary.ps1" -kompose_version ' + binaryVersion + '',
      '}}"',
    ].join(' ');
  }
}

function shell_download_kubectl_binary(p, a) {
  var binaryVersion = '<%= binaries.kubectlVersion %>';

  if (p === 'linux' || p === 'darwin') {
    return ['if [ -f dist/kubectl ]; then', 'echo "kubectl binary exists";', 'else', 'build/download_kubectl_binary.sh ' + p + ' ' + a + ' ' + binaryVersion + ';', 'fi'].join(' ');
  } else {
    return [
      'powershell -Command "& {if (Test-Path -Path "dist/kubectl.exe") {',
      'Write-Host "Skipping download, Kubectl binary exists"',
      'return',
      '} else {',
      '& ".\\build\\download_kubectl_binary.ps1" -kubectl_version ' + binaryVersion + '',
      '}}"',
    ].join(' ');
  }
}
