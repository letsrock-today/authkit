var webpack = require('webpack')

module.exports = {
  cache: true,
  entry: {
    app: './src/index.js',
    login: './src/login.js',
    vendor: './src/vendor.js',
    password: './src/password-reset.js',
  },
  output: {
    path: './dist/',
    publicPath: '/dist/',
    filename: '[name].bundle.js',
  },
  module: {
    preLoaders: [
      { test: /\.html$/, include: /src/, loader: 'riotjs', query: { type: 'none' } },
    ],
    loaders: [
      { test: /\.css$/, include: /src/, loader: 'style!css' },
      { test: /\.js$|\.html$/, include: /src/, loader: 'babel', query: { presets: 'es2015-riot' } },
      {test: /\.(png|woff|woff2|eot|ttf|svg)$/, loader: 'url-loader?limit=100000'}
    ],
  },
  babel: {
    presets: ['es2015'],
  },
  plugins: [
    new webpack.ProvidePlugin({
      riot: 'riot',
    }),
    new webpack.optimize.CommonsChunkPlugin(/* chunkName= */'vendor', /* filename= */'vendor.bundle.js'),
  ],
  devServer: {
    port: 8080,
  },
  devtool: 'source-map',
}
