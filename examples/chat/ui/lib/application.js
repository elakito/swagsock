$(function () {
    "use strict";

    var content = $('#content');
    var input = $('#input');
    var status = $('#status');
    var room = $('#room');

    var socket;
    var isopen;
    
    var author;
    var subid;
    var memid;
    var members = {};
    
    function connectWS() {
        var webSocketProtocol = "ws"
        if (window.location.protocol == "https:") {
            webSocketProtocol = "wss"
        }

        socket = swagsock.swaggersocket(webSocketProtocol + "://" + window.location.host + "/samples/chat/");
        socket.on("open", function () {
            isopen = true;
            content.append($('<p>', { text: 'Connected using websocket'}));
            input.removeAttr('disabled').focus();
            if (author === undefined) {
                status.text('Choose name:');
            } else {
                status.text(author + ': ').css('color', 'blue');
                doSubscribe(author);
                members = {};
                doMembers();
            }
        });

        socket.on("close", function () {
            content.append($('<p>', { text: 'Server closed the connection' }));
            input.attr('disabled', 'disabled');
            isopen = false;
        });

        socket.on("error", function (event) {
            content.append($('<p>', { text: 'Server encountered some problem' }));
            input.attr('disabled', 'disabled');
        });

        socket.open();
    }

    input.keydown(function(e) {
        if (e.keyCode === 13) {
            var msg = $(this).val();
            if (author === undefined) {
                author = msg;
                status.text(author + ': ').css('color', 'blue');
                doSubscribe(author);
                members = {};
                doMembers();
            } else {
                doChat(author, JSON.stringify({"text": msg}));
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

    ////////////

    function doSubscribe(name) {
        var req = socket.request().pathpattern("/v1/subscribe/{name}").pathparam("name", name).subscribe();
        socket.send(req, function (header, content) {
            if (header.code === 200) {
                var body = JSON.parse(content);
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
            }
        });
        return req.getrequestid();
    }

    function doChat(name, content) {
        var req = socket.request().pathpattern("/v1/chat/{name}").pathparam("name", name).method("POST").content(content, "application/json");
        socket.send(req);
        return req.getrequestid();
    }

    function doMembers() {
        var req = socket.request().pathpattern("/v1/members");
        socket.send(req, function (header, content) {
            var body = JSON.parse(content);
            // the initial members list
            for (var i = 0; i < body.length; i++) {
                members[body[i]] = true;
            }
            updateRoom(members);
        });
        return req.getrequestid();
    }

    ////////////
    connectWS();
});
