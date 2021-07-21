const path = require('path');
const { ProvidePlugin, IgnorePlugin } = require('webpack');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const WebpackBuildNotifierPlugin = require('webpack-build-notifier');
const CleanTerminalPlugin = require('clean-terminal-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const CleanWebpackPlugin = require('clean-webpack-plugin');
const LodashModuleReplacementPlugin = require('lodash-webpack-plugin');
const pkg = require('../package.json');
const projectRoot = path.resolve(__dirname, '..');

module.exports = {
  entry: {
    main: './app',
  },
  output: {
    filename: '[name].[hash].js',
    path: path.resolve(projectRoot, 'dist/public'),
  },
  module: {
    rules: [
      {
        test: /\.js$/,
        enforce: 'pre',
        use: [
          {
            loader: 'source-map-loader',
            options: {
              filterSourceMappingUrl: (_, resourcePath) => {
                // ignores pkgs missing sourcemaps
                return ['chardet', 'tokenize-ansi'].every((pkg) => !resourcePath.includes(pkg));
              },
            },
          },
        ],
      },
      {
        test: /\.js$/,
        exclude: /node_modules/,
        use: [
          'babel-loader',
          'auto-ngtemplate-loader',
          {
            // enforce: 'pre',
            loader: 'eslint-loader',
          },
        ],
      },
      {
        test: /\.html$/,
        exclude: path.resolve(projectRoot, './app/index.html'),
        use: [
          {
            loader: 'ngtemplate-loader',
            options: {
              relativeTo: projectRoot + '/',
            },
          },
          { loader: 'html-loader' },
        ],
      },

      {
        test: /.xml$/,
        use: 'file-loader',
      },
      {
        test: /\.css$/,
        use: [MiniCssExtractPlugin.loader, { loader: 'css-loader', options: { importLoaders: 1 } }, 'postcss-loader'],
      },
    ],
  },
  devServer: {
    contentBase: path.join(__dirname, '.tmp'),
    compress: true,
    port: 8999,
    proxy: {
      '/api': 'http://localhost:9000',
    },
    open: true,
  },
  plugins: [
    new HtmlWebpackPlugin({
      template: './app/index.html',
      templateParameters: {
        name: pkg.name,
        author: pkg.author,
      },
      manifest: './assets/ico/manifest.json',
    }),
    new WebpackBuildNotifierPlugin({
      title: 'Portainer build',
      logo: path.resolve('./assets/favicon-32x32.png'),
      suppressSuccess: true,
    }),
    new CleanTerminalPlugin(),
    new ProvidePlugin({
      $: 'jquery',
      jQuery: 'jquery',
      'window.jQuery': 'jquery',
      'window.moment': 'moment',
      moment: 'moment',
      'window.jsyaml': 'js-yaml',
      jsyaml: 'js-yaml',
    }),
    new MiniCssExtractPlugin({
      filename: '[name].[hash].css',
      chunkFilename: '[name].[id].css',
      sourceMap: true,
    }),
    new CleanWebpackPlugin(['dist/public']),
    new IgnorePlugin(/^\.\/locale$/, /moment$/),
    // new BundleAnalyzerPlugin()
    new LodashModuleReplacementPlugin({
      shorthands: true,
      collections: true,
      paths: true,
    }),
  ],
  optimization: {
    splitChunks: {
      cacheGroups: {
        vendor: {
          test: /node_modules/,
          chunks: 'initial',
          name: 'vendor',
          priority: 10,
          enforce: true,
        },
      },
    },
  },
  resolve: {
    alias: {
      Agent: path.resolve(projectRoot, 'app/agent'),
      Azure: path.resolve(projectRoot, 'app/azure'),
      Docker: path.resolve(projectRoot, 'app/docker'),
      Kubernetes: path.resolve(projectRoot, 'app/kubernetes'),
      Portainer: path.resolve(projectRoot, 'app/portainer'),
      '@': path.resolve(projectRoot, 'app'),
    },
  },
};
