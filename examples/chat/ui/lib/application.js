$(function () {
    "use strict";

    var content = $('#content');
    var input = $('#input');
    var status = $('#status');
    var room = $('#room');

    var socket;
    var isopen = false;
    var count = 0;
    
    var trackingid = createtrackingid();

    var author;
    var subid;
    var memid;
    var members = {};
    
    function connectWS() {
        var webSocketProtocol = "ws"
        if (window.location.protocol == "https:") {
            webSocketProtocol = "wss"
        }

        socket = new WebSocket(webSocketProtocol + "://" + window.location.host + "/samples/chat/?x-tracking-id="+trackingid);

        socket.onopen = function(event) {
            isopen = true;
            console.log('#### Connected using websocket');
            content.html($('<p>', { text: 'Connected using websocket'}));
            input.removeAttr('disabled').focus();
            if (author === undefined) {
                status.text('Choose name:');
            } else {
                status.text(author + ': ').css('color', 'blue');
                subid = getNextId();
                var req = JSON.stringify({ "id": subid, "method": "GET", "path": "/v1/subscribe/" + author});
                socket.send(req);
                members = {};
                memid = getNextId();
                req = JSON.stringify({ "id": memid, "method": "GET", "path": "/v1/members"});
                socket.send(req);
            }
        };

        socket.onclose = function(event) {
            content.html($('<p>', { text: 'Server closed the connection' }));
            input.attr('disabled', 'disabled');
            isopen = false;
        };

        socket.onerror = function(event) {
            content.html($('<p>', { text: 'Server encountered some problem' }));
            input.attr('disabled', 'disabled');
        };

        socket.onmessage = function(event) {
            console.log("#### onmessage " + event.data);
            var messageparts = splitMessage(event.data);
            var headers;
            try {
                headers = JSON.parse(messageparts[0]);
            } catch (e) {
                console.log('Invalid header part: ', event.data);
                return;
            }
            if (headers.heartbeat) {
                // ignore
            } else {
                var body;
                try {
                    body = JSON.parse(messageparts[1]);
                } catch (e) {
                    console.log('Invalid body part: ', event.data);
                    return;
                }
                if (headers.id === subid) {
                    // messages via the subscription or the subscription confirmation
                    if (body.type === "message") {
                        var me = body.name == author;
                        addMessage(body.name, body.text, me ? 'blue' : 'black', new Date());
                    } else if (body.type) {
                        addMessage(body.name, body.type, 'crimson', new Date());
                        if (body.type === "joined") {
                            members[body.name] = true;
                        } else if (body.type === "left") {
                            delete members[body.name];
                        }
                        updateRoom(members);
                    } else {
                        // confirmation
                    }
                } else if (headers.id === memid) {
                    // the initial members list
                    for (var i = 0; i < body.length; i++) {
                        members[body[i]] = true;
                    }
                    updateRoom(members);
                } else {
                    // messages over direct responses
                    //TODO display the messages when debug mode
                }
            }
        };
    }
    function checkWS() {
        if (!socket || socket.readyState == 3) connectWS();
    }

    input.keydown(function(e) {
        if (e.keyCode === 13) {
            var msg = $(this).val();
            if (author === undefined) {
                author = msg;
                status.text(author + ': ').css('color', 'blue');
                subid = getNextId();
                var req = JSON.stringify({ "id": subid, "method": "GET", "path": "/v1/subscribe/" + author});
                socket.send(req);
                members = {};
                memid = getNextId();
                req = JSON.stringify({ "id": memid, "method": "GET", "path": "/v1/members"});
                socket.send(req);
            } else {
                var req = JSON.stringify({ "id": getNextId(), "method": "POST", "path": "/v1/chat/" + author, "type": "application/json"})+JSON.stringify({"text": msg});
                socket.send(req);
            }
            $(this).val('');
        }
    });

    function addMessage(author, message, color, datetime) {
        content.append('<p><span style="color:' + color + '">' + author + '</span> @ ' +
                       + (datetime.getHours() < 10 ? '0' + datetime.getHours() : datetime.getHours()) + ':'
                       + (datetime.getMinutes() < 10 ? '0' + datetime.getMinutes() : datetime.getMinutes())
                       + ': ' + message + '</p>');
    }

    function updateRoom(m) {
        room.html('In room: <span style="color:green">'+Object.keys(m).join(", ")+'</span>');
    }

    // utilities
    function getNextId() {
        return (count++).toString();
    }

    function splitMessage(msg) {
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
        return [msg.substring(0, index), msg.substring(index)]
    }

    // for browser demo only as a replacement for npm uuid
    function createtrackingid() {
        const possible = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
        var text = "";
        for(var i = 0; i < 16; i++) {
            text += possible.charAt(Math.floor(Math.random() *possible.length));
        }
        return text;
    }

    connectWS();
    setInterval(checkWS, 5000);
    
});
