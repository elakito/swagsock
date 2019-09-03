(function (root, factory) {
    if (typeof exports !== 'undefined') {
        // CommonJS
        module.exports = factory();
    } else {
        // Browser globals, Window
        root.swagsock = factory();
    }
}(this, function () {
    "use strict";

    function _uuidv4() {
        if (typeof uuidv4 === 'undefined') {
            // boofa's uuid https://stackoverflow.com/a/2117523
            return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
                var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
                return v.toString(16);
            });
        } else {
            return uuidv4();
        }
    }
    
    var swagsock = function () {
        return {
            version: "0.0.1",

            swaggersocket: function (baseurl, trackingid) {
                var _wsurl;
                var _baseurl = baseurl;
                var _trackingid = trackingid ? trackingid : _uuidv4();
                var _requestid = 0;
                // requestid->callback
                var _callbacks = {};
                // requestid->issubscribed
                var _subscribed = {};
                var _ws;
                var _handlers = {};
                var _checkstatushandle;

                function getnextid () {
                    return (_requestid++).toString();
                }

                function decode (msg) {
                    var depth = 0;
                    var index = 0;
                    while (index < msg.length) {
                        var c = msg.charAt(index);
                        if (c === '{') {
                            depth++;
                        } else if (c == '}') {
                            depth--;
                        } else if (c == '\\') {
                            index++;
                        }
                        index++;
                        if (depth == 0) {
                            break;
                        }
                    }
                    return {header: msg.substring(0, index), content: msg.substring(index)};
                }

                function callhandler(name, arg) {
                    var handler = _handlers[name];
                    if (handler) {
                        handler(arg);
                    }
                }

                function checkstatus() {
                    if (!_ws || _ws.readyState === 3) connect();
                }
                
                function connect() {
                    _ws = new WebSocket(_wsurl);
                    _ws.onopen = function (event) {
                        callhandler("open");
                    };
                    _ws.onclose = function (event) {
                        // TODO need another way to inform the client app about the broken connection.
                        _subscribed = {};
                        for (var cid in _callbacks) {
                            _callbacks[cid]({"id": cid, "code":503}, "")
                            delete _callbacks[cid];
                        }
                        callhandler("close");
                    };
                    _ws.onerror = function (event) {
                        callhandler("error", event.data);
                    };
                    _ws.onmessage = function (event) {
                        var message = event.data;
                        var decoded = decode(message);
                        var header;
                        var content;
                        try {
                            header = JSON.parse(decoded.header);
                            content = decoded.content;
                            if (header.heartbeat) {
                                // ignore
                            } else if (header.id) {
                                var callback = _callbacks[header.id];
                                if (callback) {
                                    callback(header, content);
                                    if (!(header.id in _subscribed)) {
                                        delete _callbacks[header.id];
                                    }
                                }
                            } else {
                                // no id supplied in the response, so just log the plain result
                                console.log("res*" + ":", message);
                            }
                        } catch (e) {
                            console.log('Invalid response: ', message + "; error="+e);
                        }
                    };
                }
                
                return {
                    open: function () {
                        if (_baseurl.indexOf("http:") === 0) {
                            _wsurl = "ws" + _baseurl.substring(4);
                        } else if (_baseurl.indexOf("ws:") === 0) {
                            _wsurl = _baseurl;
                        }
                        
                        if (_wsurl.indexOf("?") > 0) {
                            _wsurl += "&x-tracking-id=" + _trackingid;
                        } else {
                            _wsurl += "?x-tracking-id=" + _trackingid;
                        }
                        connect();
                        _checkstatushandle = setInterval(checkstatus, 5000);
                    },
                    close: function () {
                        _ws.close();
                    },
                    on: function(event, handler) {
                        // open, close, error
                        _handlers[event] = handler;
                    },
                    send: function (req, callback) {
                        var reqid = getnextid();
                        req._requestid = reqid;

                        if (callback) {
                            _callbacks[reqid] = callback;
                        }
                        if (req._subscribe) {
                            _subscribed[reqid] = true;
                        } else if (req._unsubscribe) {
                            delete _subscribed[req._unsubscribe];
                            delete _callbacks[req._unsubscribe];
                        }
                        var msg = {"id":reqid,"method":req.getmethod(),"path":req.getpath()};
                        var content = "";
                        if (req.getcontent()) {
                            msg.type = req.gettype();
                            content = req.getcontent();
                        }
                        var smsg = JSON.stringify(msg)+content;
                        _ws.send(smsg);
                    },
                    request: function () {
                        return {
                            _requestid: undefined,
                            _method: "GET",
                            _pathpattern: undefined,
                            _pathparams: {},
                            _queryparams: {},
                            _headerparams: {},
                            _content: undefined,
                            _type: undefined,
                            _subscribe: undefined,                            
                            _unsubscribe: undefined,

                            method: function (method) {
                                this._method = method;
                                return this;
                            },
                            pathpattern: function (pathpattern) {
                                this._pathpattern = pathpattern;
                                return this;
                            },
                            pathparam: function (name, value) {
                                this._pathparams[name] = value;
                                return this;
                            },
                            queryparam: function (name, value) {
                                this._queryparams[name] = value;
                                return this;
                            },
                            headerparam: function (name, value) {
                                this._headerparams[name] = value;
                                return this;
                            },
                            content: function (cvalue, ctype) {
                                this._content = cvalue;
                                this._type = ctype;
                                return this;
                            },
                            subscribe: function () {
                                this._subscribe = true;
                                return this;
                            },
                            unsubscribe: function (sid) {
                                this._unsubscribe = sid;
                                return this;
                            },
                            getrequestid: function () {
                                return this._requestid;
                            },
                            getmethod: function () {
                                return this._method;
                            },
                            getpath: function () {
                                //TODO make the path building code more efficient
                                var path = this._pathpattern;
                                for (var pp in this._pathparams) {
                                    path = path.replace("{"+pp+"}", encodeURIComponent(this._pathparams[pp]));
                                }
                                var querystr = "";
                                for (var qp in this._queryparams) {
                                    if (querystr.length === 0) querystr = "?";
                                    querystr += encodeURIComponent(qp) + "=" + encodeURIComponent(this._queryparams[qp]);
                                }
                                return path + querystr;
                            },
                            getcontent: function () {
                                return this._content;
                            },
                            gettype: function () {
                                return this._type;
                            }
                        };
                    }
                };
            }
        };
    }();
    return swagsock;
}));

