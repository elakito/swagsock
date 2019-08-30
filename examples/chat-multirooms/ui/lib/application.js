$(function () {
    "use strict";

    var content = $('#content');
    var checklist = $('#checklist');
    var createroom = $('#createroom');
    var input = $('#input');
    var status = $('#status');
    var inroom = $('#inroom');

    var buttonroom = $('#buttonroom');

    var socket;
    var isopen = false;
    var count = 0;
    
    var trackingid = createtrackingid();

    // chat-author
    var author;
    // subid->room
    var subids = {};
    // request-id for op rooms
    var roomsid;
    // room -> subid if subscribed otherwise ''
    var rooms = {};
    // memsid -> room
    var memids = {};
    // room in focus
    var chatroom;
    // room -> (member -> 
    var members = {};
    // room -> 
    var contents = {};

    
    checklist.hide();
    createroom.hide();
    
    function connectWS() {
        var webSocketProtocol = "ws"
        if (window.location.protocol == "https:") {
            webSocketProtocol = "wss"
        }

        socket = new WebSocket(webSocketProtocol + "://" + window.location.host + "/samples/chat-multirooms/?x-tracking-id="+trackingid);

        socket.onopen = function(event) {
            isopen = true;
            content.html($('<p>', { text: 'Connected using websocket'}));
            input.removeAttr('disabled').focus();
            if (author === undefined) {
                status.text('Choose name:');
            } else {
                status.text(author + ': ').css('color', 'blue');
                rooms = {};
                members = {};
                roomsid = getNextId();
                var req = JSON.stringify({ "id": roomsid, "method": "GET", "path": "/v1/rooms"});
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
                if (headers.id in subids) {
                    // messages via the subscriptions or the subscription confirmation
                    var subid = headers.id;
                    if (body.type === undefined) {
                        // subscription confirmation
                    } else if (subid === rooms[body.room]) {
                        if (body.type === "message") {
                            addMessage(body.name, body.room, body.text, (body.name === author) ? 'blue' : 'black', new Date());
                        } else if (body.type === "joined") {
                            addMessage(body.name, body.room, body.type, 'crimson', new Date());
                            if (body.name === author) {
                                var memid = getNextId();
                                memids[memid] = body.room;
                                var req = JSON.stringify({ "id": memid, "method": "GET", "path": "/v1/rooms/" + body.room});
                                socket.send(req);
                            } else {
                                members[body.room][body.name] = "";
                                if (chatroom === body.room) {
                                    updateInRoom(members[body.room]);
                                }
                            }
                        } else if (body.type === "left") {
                            addMessage(body.name, body.room, body.type, 'crimson', new Date());
                            if (body.name === author) {
                                delete members[body.room];
                                rooms[body.room] = "";
                                updateInRoom(undefined);
                            } else {
                                delete members[body.room][body.name];
                                if (chatroom === body.room) {
                                    updateInRoom(members[body.room]);
                                }
                            }
                        } else {
                            // confirmation
                        }
                    } else {
                        if (body.type === "joined" && !(body.room in rooms)) {
                            rooms[body.room] = "";
                            addCheckbox(body.room, false);
                        }
                    }
                } else if (headers.id in memids) {
                    var room = memids[headers.id];
                    var mems = {};
                    for (var i = 0; i < body.length; i++) {
                        mems[body[i]] = "";
                    }
                    members[room] = mems;
                    if (chatroom === room) {
                        updateInRoom(mems);
                    }
                } else if (headers.id === roomsid) {
                    // the initial rooms list
                    for (var i = 0; i < body.length; i++) {
                        rooms[body[i]] = "";
                    }
                    createCheckboxList();
                    // subscribe automatically to the general room
                    addCheckbox("general", true);
                    $('#checkgeneral').attr('disabled', 'disabled');
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

    buttonroom.click(function () {
        showModal("New room", "Enter a new room name", "untitled", newRoom);
    });


    function newRoom() {
        var room = $('#updateinput').val();
        if (room in rooms) {
            return;
        }
        addCheckbox(room, true);
    }

    input.keydown(function(e) {
        if (e.keyCode === 13) {
            var msg = $(this).val();
            if (author === undefined) {
                author = msg;
                status.text(author + ': ').css('color', 'blue');
                rooms = {};
                roomsid = getNextId();
                var req = JSON.stringify({ "id": roomsid, "method": "GET", "path": "/v1/rooms"});
                console.log("### sending req="+req);
                socket.send(req);
                $(this).val('');

                checklist.show();
                createroom.show();
            } else {
                var active = $(".tab-pane.active").attr("id");
                if (active === undefined) {
                    alert("you are not in a room. choose a room or create a new one.");
                } else {
                    var room = active.substring(4);
                    if (rooms[room].length == 0) {
                        alert("you are not in a room. choose a room or create a new one.");
                    } else {
                        var req = JSON.stringify({ "id": getNextId(), "method": "POST", "path": "/v1/chat/" + author + "/" + room, "type": "application/json"})+JSON.stringify({"text": msg});
                        socket.send(req);
                        $(this).val('');
                    }
                }
            }
        }
    });

    function addMessage(author, room, message, color, datetime) {
        $('#room'+room).append('<p><span style="color:' + color + '">' + author + '</span> @ ' +
                       + (datetime.getHours() < 10 ? '0' + datetime.getHours() : datetime.getHours()) + ':'
                       + (datetime.getMinutes() < 10 ? '0' + datetime.getMinutes() : datetime.getMinutes())
                       + ': ' + message + '</p>');
    }

    function addCheckbox(name, c) {
        if ($('#check'+name).val() === undefined) {
            $('#checklist').append('<label class="checkbox-inline"><input type="checkbox" id="check'+name+'">'+name+"</label>");

            $('#check'+name).click(function() {
                var room = $(this).attr('id').substring(5);
                if ($(this).is(':checked')) {
                    if ($('#navroom'+room).val() === undefined) {
                        $('.nav-tabs').append('<li><a data-toggle="tab" href="#room'+room+'" id="navroom'+room+'">'+room+'</a></li>');
                        $('.tab-content').append('<div id="room'+room+'" class="tab-pane fade"></div>');
                        $('.nav-tabs a').on('shown.bs.tab', function(event) {
                            chatroom = $(event.target).text();
                            var mems = members[chatroom];
                            if (mems !== undefined) {
                                updateInRoom(mems);
                            }
                        });
                    }
                    chatroom = room;
                    $('#navroom'+room).tab('show');
                    var subid = getNextId();
                    rooms[room] = subid;
                    subids[subid] = room;
                    var req = JSON.stringify({ "id": subid, "method": "GET", "path": "/v1/subscribe/" + author + "/" + room});
                    console.log("#### sending req "+req);
                    socket.send(req);
                } else {
                    var subid = rooms[room];
                    var req = JSON.stringify({ "id": getNextId(), "method": "DELETE", "path": "/v1/unsubscribe/" +subid});
                    console.log("#### sending req "+req);
                    socket.send(req);
                }
            });
        }

        if (c) {
            $('#check'+name).trigger('click');
        }
    }

    function createCheckboxList() {
        for (var name in rooms) {
            addCheckbox(name, false);
        }
    }
    
    function updateInRoom(m) {
        var val;
        if (m) {
            val = Object.keys(m).join(", ");
        } else {
            val = "?";
        }
        inroom.html('In room: <span style="color:green">'+val+'</span>');
    }

    function showModal(title, desc, value, action) {
        $("#updatemodal").on('shown.bs.modal', 
                             function (e){
                                 $(".modal-dialog #updatetitle").text(title);
                                 $(".modal-dialog #updatename").text(desc);
                                 $(".modal-dialog #updateinput").val(value);
                                 $(".modal-dialog #updatebutton").off('click');
                                 $(".modal-dialog #updatebutton").click(action);
                             });
        $("#updatemodal").modal("show")
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
