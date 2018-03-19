const HtmlWebpackPlugin = require('html-webpack-plugin');
const path = require('path');

module.exports = {
  entry: ['./src/js/index.js',
  'jquery'
  ],
  output: {
    filename: 'bundle.js',
    path: path.resolve(__dirname, 'dist')
  },
  externals: {
    jquery: 'jQuery'
  },
  module: {
    rules: [
      {
        test: /\.css$/,
        use: [
          'style-loader',
          'css-loader'
        ]
      }
    ]
  },
  devServer: {
    contentBase: './dist'
  },
  plugins: [
      new HtmlWebpackPlugin({template: './src/templates/index.html',
                             filename: 'index.html'}),
      new HtmlWebpackPlugin({template: './src/templates/product.html',
                             filename: 'product.html'}),
      new HtmlWebpackPlugin({template: './src/templates/search.html',
                             filename: 'search.html'}),
      new HtmlWebpackPlugin({template: './src/templates/thank_you.html',
                             filename: 'thank_you.html'}),
      new HtmlWebpackPlugin({template: './src/templates/user_orders.html',
                             filename: 'user_orders.html'}),
      new HtmlWebpackPlugin({template: './src/templates/user_profile.html',
                             filename: 'user_profile.html'})
  ]
};

