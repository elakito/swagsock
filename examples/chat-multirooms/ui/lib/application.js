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
    var isopen;
    
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
    // room -> (member -> "")
    var members = {};
    
    checklist.hide();
    createroom.hide();
    
    function connectWS() {
        var webSocketProtocol = "ws"
        if (window.location.protocol == "https:") {
            webSocketProtocol = "wss"
        }

        socket = swagsock.swaggersocket(webSocketProtocol + "://" + window.location.host + "/samples/chat-multirooms/");
        socket.on("open", function () {
            isopen = true;
            content.html($('<p>', { text: 'Connected using websocket'}));
            for (var subid in subids) {
                addStatusMessage(subids[subid], 'Connected using websocket');
            }
            input.removeAttr('disabled').focus();
            if (author === undefined) {
                status.text('Choose name:');
            } else {
                status.text(author + ': ').css('color', 'blue');
                rooms = {};
                members = {};
                doRooms();
                var oldsubids = subids;
                subids = {};
                for (var oldsubid in oldsubids) {
                    // subscribe to the previously subscribed rooms
                    var oldroom = oldsubids[oldsubid]
                    var subid = doSubscribe(author, oldroom);
                    rooms[oldroom] = subid;
                    subids[subid] = oldroom;
                }
            }
        });
        
        socket.on("close", function () {
            content.html($('<p>', { text: 'Server closed the connection' }));
            for (var subid in subids) {
                addStatusMessage(subids[subid], 'Server closed the connection');
            }
            input.attr('disabled', 'disabled');
            isopen = false;
        });

        socket.on("error", function (event) {
            content.html($('<p>', { text: 'Server encountered some problem' }));
            input.attr('disabled', 'disabled');
        });

        socket.open();
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
                doRooms();
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
                        doChat(author, room, JSON.stringify({"text": msg}));
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

    function addStatusMessage(room, message) {
        $('#room'+room).append($('<p>', { text: message }));
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
                    var subid = doSubscribe(author, room);
                    rooms[room] = subid;
                    subids[subid] = room;
                } else {
                    var subid = rooms[room];
                    doUnsubscribe(subid);
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

    ////////////

    function doSubscribe(name, room) {
        var req = socket.request().pathpattern("/v1/subscribe/{name}/{room}").pathparam("name", name).pathparam("room", room).subscribe();
        socket.send(req, function (header, content){
            if (header.code === 200) {
                // messages via the subscriptions or the subscription confirmation
                var subid = header.id;
                var body = JSON.parse(content);
                if (body.type === undefined) {
                    // subscription confirmation
                } else if (subid === rooms[body.room]) {
                    if (body.type === "message") {
                        addMessage(body.name, body.room, body.text, (body.name === author) ? 'blue' : 'black', new Date());
                    } else if (body.type === "joined") {
                        addMessage(body.name, body.room, body.type, 'crimson', new Date());
                        if (body.name === author) {
                            var memid = doRoomMembers(body.room);
                            memids[memid] = body.room;
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
            }
        });
        return req.getrequestid();
    }

    function doUnsubscribe(sid) {
        var req = socket.request().pathpattern("/v1/unsubscribe/{sid}").pathparam("sid", sid).method("DELETE").unsubscribe(sid);
        socket.send(req);
        return req.getrequestid();
    }

    function doChat(name, room, content) {
        var req = socket.request().pathpattern("/v1/chat/{name}/{room}").pathparam("name", name).pathparam("room", room).method("POST").content(content, "application/json");
        socket.send(req);
        return req.getrequestid();
    }

    function doRooms() {
        var req = socket.request().pathpattern("/v1/rooms");
        socket.send(req, function (header, content) {
            if (header.code === 200) {
                var body = JSON.parse(content);
                // the initial rooms list
                for (var i = 0; i < body.length; i++) {
                    var room = body[i];
                    if (!rooms[room]) {
                        rooms[room] = "";
                    }
                }
                createCheckboxList();
                // subscribe automatically to the general room
                addCheckbox("general", true);
                $('#checkgeneral').attr('disabled', 'disabled');
            }
        });
        return req.getrequestid();
    }

    function doRoomMembers(room) {
        var req = socket.request().pathpattern("/v1/rooms/{room}").pathparam("room", room);
        socket.send(req, function (header, content) {
            if (header.code === 200) {
                var room = memids[header.id];
                var body = JSON.parse(content);
                var mems = {};
                for (var i = 0; i < body.length; i++) {
                    mems[body[i]] = "";
                }
                members[room] = mems;
                if (chatroom === room) {
                    updateInRoom(mems);
                }
            }
        });
        return req.getrequestid();
    }

    ////////////
    connectWS();
});
