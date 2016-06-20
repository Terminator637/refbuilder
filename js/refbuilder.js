if (window.refbuilder===undefined) {
refbuilder = {}
refbuilder.idx = {}
refbuilder.onload = []
refbuilder.getRoot = function() {
    if (this.root === undefined) {
        this.root = document.getElementById('ref_script').src.slice(0,-"assets/refbuilder.js".length)
    }
    return this.root
}
refbuilder.currentId = function() {
    if (this.id === undefined) {
        var id = document.location.toString().slice(this.getRoot().length)
        var ih = "index.html"
        if (id.slice(-ih.length)==ih)
            id = id.slice(0,-ih.length)
        if (id.slice(-1)=="/")
            id = id.slice(0,-1)
        this.id = id
    }
    return this.id
}

refbuilder.runScript = function(script, cb) {
    var tag = document.createElement('script')
    tag.type = 'text/javascript'
    tag.async = true
    if (cb !== undefined)
        tag.onload = cb
    tag.src = script
    document.getElementsByTagName('head')[0].appendChild(tag);
}
refbuilder.addStylesheet = function(script, cb) {
    var tag = document.createElement('link')
    tag.rel = 'stylesheet'
    tag.async = true
    if (cb !== undefined)
        tag.onload = cb
    tag.href = script
    document.getElementsByTagName('head')[0].appendChild(tag);
}

refbuilder.shallRun = function() {
    try {
        if (window.self !== window.top)
            return false
        if (window.sessionStorage.getItem("refbuilder_frames"))
            return false
        return true
    } catch (e) {
        return false
    }
}

if (refbuilder.shallRun()) {
    refbuilder.runScript(refbuilder.getRoot()+'assets/jquery.min.js', function () {
        var tag = document.createElement('div')
        tag.id = 'ref_navigation'
        tag.innerHTML='<input id="ref_tab1" type="radio" name="ref_tabs" checked>' +
                      '<label for="ref_tab1" title="Contents">Contents</label>'+
                      '<input id="ref_tab2" type="radio" name="ref_tabs">' +
                      '<label for="ref_tab2" title="Index">Index</label>'
        var tag3 = document.createElement('div')
        tag3.id = 'ref_structure'
        tag.appendChild(tag3)
        var tag4 = document.createElement('div')
        tag4.id = 'ref_index'
        tag4.innerHTML='<form id="ref_idx_form"><input id="ref_idx_input" name="ref_idx_input" type="search"/><button id="ref_idx_search" type="submit">Search</button></form><ul id="ref_idx_results">'
        tag.appendChild(tag4)
        var tag2 = document.createElement('iframe')
        tag2.id = 'ref_content'
        tag2.sandbox = "allow-same-origin"
        var fubar=false;
        tag2.addEventListener('load', function (event) {
            var loc = refbuilder.tag_content.contentWindow.location.toString()
            if (loc.slice(-1)=="?") loc=loc.slice(0,-1)
            document.title = refbuilder.tag_content.contentDocument.title
            if (fubar)
                fubar=false
            else
                history.pushState(loc,'',loc)
        })
        window.addEventListener('popstate', function (event) {
            fubar=true;
            tag2.src = event.state+"?"
        })
        tag2.src = window.location.toString()+"?"
        var body = document.getElementsByTagName('body')[0]
        while (body.firstChild) {
            body.removeChild(body.firstChild)
        }
        body.appendChild(tag)
        body.appendChild(tag2)
        refbuilder.tag_structure = tag3
        refbuilder.tag_content = tag2
        refbuilder.addStylesheet(refbuilder.getRoot()+'assets/style.min.css', function () {
            refbuilder.addStylesheet(refbuilder.getRoot()+'assets/refbuilder.css')
        })
        refbuilder.runScript(refbuilder.getRoot()+'assets/jstree.min.js', function() {
            $.vakata.storage = {
                set : function (key, val) { return window.sessionStorage.setItem(key, val) },
                get : function (key) { return window.sessionStorage.getItem(key) },
                del : function (key) { return window.sessionStorage.removeItem(key) }
            }
            if (refbuilder.shallRun()) {
                refbuilder.runScript(refbuilder.getRoot()+'tree.jsonp')
                refbuilder.runScript(refbuilder.getRoot()+'idx.jsonp')
                refbuilder.runScript(refbuilder.getRoot()+'assets/Snowball.min.js', function () {
                    refbuilder.stemmer = new Snowball("russian")
                    refbuilder.stem = function (word) {
                        refbuilder.stemmer.setCurrent(word)
                        refbuilder.stemmer.stem()
                        return refbuilder.stemmer.getCurrent()
                    }
                    document.getElementById('ref_idx_form').addEventListener('submit', function (event) {
                        event.preventDefault()
                        var words = document.getElementById('ref_idx_input').value.split(" ")
                        console.log(words)
                        var result
                        for (var i=0;i<words.length;i++) {
                            var word = refbuilder.stem(words[i])
                            var files = refbuilder.idx[word]
                            if (files) {
                                if (result===undefined) {
                                    result={}
                                    for (k in files) {
                                        result[k] = files[k]
                                    }
                                } else {
                                    var toremove=[]
                                    for (k in result) {
                                        if (!files[k]) {
                                            toremove.push(k)
                                        }
                                    }
                                    for (var j in toremove) {
                                        console.log('Removing',toremove[j])
                                        delete result[toremove[j]]
                                    }
                                }
                            } else {
                                result={}
                            }
                        }
                        var output = document.getElementById('ref_idx_results')
                        while (output.firstChild) {
                            output.removeChild(output.firstChild)
                        }
                        for (k in result) {
                            var option = document.createElement('li')
                            var treenode = refbuilder.tree.get_node('ref_node_'+k)
                            var link = document.createElement('a')
                            link.href=refbuilder.getRoot()+k
                            link.title=k
                            link.addEventListener('click', function (event) {
                                event.preventDefault()
                                console.log(event.target.title)
                                var treenode=refbuilder.tree.get_node('ref_node_'+event.target.title)
                                refbuilder.tree.activate_node(treenode)
                            })
                            if (treenode) {
                                link.appendChild(document.createTextNode(treenode.text))
                            } else {
                                link.appendChild(document.createTextNode(k))
                            }
                            option.appendChild(link)
                            output.appendChild(option)
                        }
                        return false
                    })
                })
            }
        })
    })
}
refbuilder.open_subtree = function(id, cb, start) {
    if (!start) start=0;
    var delim=id.slice(start).indexOf("/")
    var part;
    if (delim<0) {
        part = id
    } else {
        part = id.slice(0,delim+start)
    }
    //console.log("id",part)
    //console.log($("#ref_node_"+part))
    refbuilder.tree.open_node(document.getElementById("ref_node_"+part), function () {
        if (delim>=0) {
            start+=delim+1
            refbuilder.open_subtree(id, cb, start)
        } else {
            if (cb) cb()
        }
    })
}

refbuilder.load_tree = function(tree_data) {
    $(function() {

        refbuilder.tree = $(refbuilder.tag_structure).jstree({
            "core" : {
                "animation" : 0,
                "check_callback" : true,
                'data' : tree_data,
            },
            "types" : {
                "#" : {
                    "max_children" : 1,
                    "valid_children" : ["root"]
                },
                "root" : {
                    "icon" : "ref_icon_root",
                    "valid_children" : ["default","file"]
                },
                "default" : {
                    "icon" : "ref_icon_folder",
                    "valid_children" : ["default","file"]
                },
                "file" : {
                    "icon" : "ref_icon_file",
                    "valid_children" : []
                }
            },
            "plugins" : [
                "state", "types"
            ]
        }).jstree(true)
        for (var i=0;i<refbuilder.onload.length;i++) {
            try {
                refbuilder.onload[i]()
            } catch(e) {
                alert(e)
            }
        }
        $('#ref_structure').on('ready.jstree', function() {
            var id=refbuilder.currentId()
            refbuilder.tree.open_node($("#ref_root"), function () {
                refbuilder.open_subtree(id, function() {
                    refbuilder.tree.deselect_all()
                    refbuilder.tree.select_node(document.getElementById("ref_node_"+id))
                })
            })
        })
        $('#ref_structure').on('activate_node.jstree', function(crap,event) {
            if (event.node && event.node.id) {
                var id = event.node.id;
                if (id=="ref_root")
                    id = ""
                else
                    id = id.slice('ref_node_'.length)
                if (event.node.type=="default")
                    id = id+"/index.html"
                //ajax.get    
                refbuilder.tag_content.src = refbuilder.getRoot()+id+"?"
            }
        })

    })
}

refbuilder.load_idx = function(idx_data) {
    refbuilder.idx = idx_data
}

}
