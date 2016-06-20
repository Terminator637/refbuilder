#!/bin/bash
set -xeu -o pipefail
mkdir -p assets
#cp deps/bower_components/jquery/dist/jquery.min.js deps/bower_components/jstree/dist/jstree.min.js deps/bower_components/jstree/dist/themes/default/style.min.css assets/
pushd deps
test -d node_modules || npm i
popd

cp deps/node_modules/jquery/dist/jquery.min.js assets/

if [ -f compiler.jar ]; then
    java -jar compiler.jar --js_output_file=assets/jstree.min.js deps/node_modules/jstree/src/{jstree.js,jstree.state.js,jstree.types.js,vakata-jstree.js}
else
    cp deps/node_modules/jstree/dist/jstree.min.js assets/
fi
cp deps/node_modules/jstree/dist/themes/default/style.min.css assets/

cp deps/node_modules/snowball/stemmer/lib/Snowball.min.js assets/

cp img/icons.png img/throbber.gif css/*.css assets/

if [ -f compiler.jar ]; then
    java -jar compiler.jar --js_output_file=assets/refbuilder.js js/*.js
else
    cp js/*.js assets/
fi

pushd assets
find -mindepth 1 '!' -iname index |sed 's!^./!!' > index
popd
