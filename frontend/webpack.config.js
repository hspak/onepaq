 var path = require('path');

module.exports = {
  entry: './index.js',
  output: {
    filename: 'bundle.js',
    path: path.resolve(__dirname, "..", "public", "js"),
  },
  module: {
    rules: [{
      test: /\.jsx?$/,
      loader: 'babel-loader',
      include: [
        path.resolve(__dirname)
      ],
      exclude: [
        path.resolve(__dirname, "node_modules"),
      ],
      options: {
        presets: [
          'react',
          'env'
        ]
      }
    }]
  },
  resolve: {
    extensions: [".jsx", ".js"]
  }
}
