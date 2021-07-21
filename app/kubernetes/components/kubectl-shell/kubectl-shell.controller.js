import { Terminal } from 'xterm';
import * as fit from 'xterm/lib/addons/fit/fit';

export default class KubectlShellController {
  /* @ngInject */
  constructor(TerminalWindow, $window, $async, EndpointProvider, LocalStorage, KubernetesConfigService) {
    this.$async = $async;
    this.$window = $window;
    this.TerminalWindow = TerminalWindow;
    this.EndpointProvider = EndpointProvider;
    this.LocalStorage = LocalStorage;
    this.KubernetesConfigService = KubernetesConfigService;
  }

  disconnect() {
    this.state.checked = false;
    this.state.icon = 'fas fa-window-minimize';
    this.state.shell.socket.close();
    this.state.shell.term.dispose();
    this.state.shell.connected = false;
    this.TerminalWindow.terminalclose();
  }

  screenClear() {
    this.state.shell.term.clear();
  }

  miniRestore() {
    if (this.state.css === 'mini') {
      this.state.css = 'normal';
      this.state.icon = 'fas fa-window-minimize';
      this.TerminalWindow.terminalopen();
    } else {
      this.state.css = 'mini';
      this.state.icon = 'fas fa-window-restore';
      this.TerminalWindow.terminalclose();
    }
  }

  configureSocketAndTerminal(socket, term) {
    socket.onopen = function () {
      const terminal_container = document.getElementById('terminal-container');
      term.open(terminal_container);
      term.setOption('cursorBlink', true);
      term.focus();
      term.fit();
      term.writeln('#Run kubectl commands inside here');
      term.writeln('#e.g. kubectl get all');
      term.writeln('');
    };

    term.on('data', function (data) {
      socket.send(data);
    });

    this.$window.onresize = function () {
      term.fit();
    };

    socket.onmessage = function (msg) {
      term.write(msg.data);
    };

    socket.onerror = function (err) {
      this.disconnect();
      this.Notifications.error('Failure', err, 'Websocket connection error');
    }.bind(this);

    this.state.shell.socket.onclose = this.disconnect.bind(this);

    this.state.shell.connected = true;
  }

  connectConsole() {
    this.TerminalWindow.terminalopen();
    this.state.checked = true;
    this.state.css = 'normal';

    const params = {
      token: this.LocalStorage.getJWT(),
      endpointId: this.EndpointProvider.endpointID(),
    };

    const wsProtocol = this.state.isHTTPS ? 'wss://' : 'ws://';
    const path = '/api/websocket/kubernetes-shell';
    const queryParams = Object.entries(params)
      .map(([k, v]) => `${k}=${v}`)
      .join('&');
    const url = `${wsProtocol}${window.location.host}${path}?${queryParams}`;

    Terminal.applyAddon(fit);
    this.state.shell.socket = new WebSocket(url);
    this.state.shell.term = new Terminal();

    this.configureSocketAndTerminal(this.state.shell.socket, this.state.shell.term);
  }

  async downloadKubeconfig() {
    await this.KubernetesConfigService.downloadConfig();
  }

  $onInit() {
    return this.$async(async () => {
      this.state = {
        css: 'normal',
        checked: false,
        icon: 'fa-window-minimize',
        isHTTPS: this.$window.location.protocol === 'https:',
        shell: {
          connected: false,
          socket: null,
          term: null,
        },
      };
    });
  }

  $onDestroy() {
    if (this.state.shell.connected) {
      this.disconnect();
      this.$window.onresize = null;
    }
  }
}
