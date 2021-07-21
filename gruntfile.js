var os = require('os');
var loadGruntTasks = require('load-grunt-tasks');
const webpackDevConfig = require('./webpack/webpack.develop');
const webpackProdConfig = require('./webpack/webpack.production');

var arch = os.arch();
if (arch === 'x64') arch = 'amd64';

var portainer_data = '${PORTAINER_DATA:-/tmp/portainer}';

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
      dockerLinuxComposeVersion: '1.27.4',
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

  grunt.task.registerTask('devopsbuild', 'devopsbuild:<platform>:<arch>:env', function (p, a, env = 'prod') {
    grunt.task.run([
      'config:prod',
      `env:${env}`,
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
  testing: {
    NODE_ENV: 'testing',
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
  return 'build/build_binary_azuredevops.sh ' + p + ' ' + a + ';';
}

function shell_run_container() {
  return [
    'docker rm -f portainer',
    'docker run -d -p 8000:8000 -p 9000:9000 -v $(pwd)/dist:/app -v ' +
      portainer_data +
      ':/data -v /var/run/docker.sock:/var/run/docker.sock:z -v /var/run/docker.sock:/var/run/alternative.sock:z -v /tmp:/tmp --name portainer portainer/base /app/portainer',
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
  
  return [
    'if [ -f dist/docker ] || [ -f dist/docker.exe ]; then',
    'echo "docker binary exists";',
    'else',
    'build/download_docker_binary.sh ' + ip + ' ' + ia + ' ' + binaryVersion + ';',
    'fi',
  ].join(' ');
}

function shell_download_docker_compose_binary(p, a) {
  var ps = { windows: 'win', darwin: 'mac' };
  var as = { arm: 'armhf', arm64: 'aarch64' };
  var ip = ps[p] || p;
  var ia = as[a] || a;
  var binaryVersion = p === 'windows' ? '<%= binaries.dockerWindowsComposeVersion %>' : '<%= binaries.dockerLinuxComposeVersion %>';
  
  return [
    'if [ -f dist/docker-compose ] || [ -f dist/docker-compose.exe ]; then',
    'echo "Docker Compose binary exists";',
    'else',
    'build/download_docker_compose_binary.sh ' + ip + ' ' + ia + ' ' + binaryVersion + ';',
    'fi',
  ].join(' ');
}

function shell_download_kompose_binary(p, a) {
  var binaryVersion = '<%= binaries.komposeVersion %>';
  
  return [
    'if [ -f dist/kompose ] || [ -f dist/kompose.exe ]; then',
    'echo "kompose binary exists";',
    'else',
    'build/download_kompose_binary.sh ' + p + ' ' + a + ' ' + binaryVersion + ';',
    'fi',
  ].join(' ');
}

function shell_download_kubectl_binary(p, a) {
  var binaryVersion = '<%= binaries.kubectlVersion %>';
  
  return [
    'if [ -f dist/kubectl ] || [ -f dist/kubectl.exe ]; then',
    'echo "kubectl binary exists";',
    'else',
    'build/download_kubectl_binary.sh ' + p + ' ' + a + ' ' + binaryVersion + ';',
    'fi',
  ].join(' ');
}
