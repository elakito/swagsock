/**
 * client.js
 * 
 * An atmosphere node client program to interactively calls the greeter service
 * 
 * Usages: 
 * node client.js [options] [greeter-url]
 *
 * Options:
 *   -v    enable verbose mode
 *
 * Usage Samples:
 * node client.js http://localhost:8091/samples/greeter
 * 
 */

"use strict";

const HOST_URL = 'http://localhost:8091/samples/greeter';
const PROTOCOL_VERSION = "2.0";

var trace = false;

var hosturl = HOST_URL;

const uuidv4 = require('uuid/v4');
var clientid = uuidv4();

var arg;
for (var i = 2; i < process.argv.length; i++) {
    arg = process.argv[i];
    if (arg === "-v") {
        trace = true;
        arg = undefined;
    }
}
if (arg != undefined) {
    hosturl = arg;
}

console.log("Client ID: " + clientid);
console.log("Host URL: " + hosturl)

var reader = require('readline');
var prompt = reader.createInterface(process.stdin, process.stdout);

var atmosphere = require('atmosphere.js');

var request = { url: hosturl,
                transport : 'websocket',
                enableProtocol: false,
                trackMessageLength: false,
                reconnectInterval : 5000,
                uuid: clientid};
var isopen = false;

// request/respons index
var count = 0;

// greeter's consle commands
const COMMAND_LIST = 
    [["ping",       "Ping", "ping"],
     ["echo",       "Echo the message", "echo text"],
     ["greet",      "Greet the person", "greet name text"],
     ["status",     "Display the greeting status", "status name"],
     ["summary",    "Display the greeting summary", "summary"],
     ["subscribe",  "Subscribe to the greeting events", "subscribe"],
     ["unsubscribe","Unsubscribe from the greeting events", "unsubscribe sid"],
     ["quit",       "Quit the application", "quit"]];


/////////////// utilities

function getNextId() {
    return (count++).toString();
}

function selectOption(c, opts) {
    var i = c.length == 0 ? 0 : parseInt(c);
    if (!(i >= 0 && i < opts.length)) {
        console.log('Invalid selection: ' + c + '; Using ' + opts[0]);
        i = 0;
    }
    return opts[i];
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

function getArgs(name, msg, num) {
    var sp = name.length;
    if (msg.length > name.length && msg.charAt(name.length) != ' ') {
        // remove the command suffix
        sp = msg.indexOf(' ', name.length);
        if (sp < 0) {
            sp = msg.length;
        }
    }
    var v = msg.substring(sp).trim();
    if (v.length > 0) {
        if (num == 0) {
            // return as a single value
            return v;
        } else {
            // return as an array of num + 1 elements containing the first num arguments and the rest
            var params = [];
            while (num > 0) {
                var param;
                if (num == 1) {
                    params.push(v);
                } else {
                    var p = v.indexOf(' ', 0);
                    if (p < 0) {
                        break;
                    }
                    param = v.substring(0, p);
                    v = v.substring(p + 1).trim();
                }
                params.push(param);
                num--;
            }
            if (num > 1) {
                throw "Invalid arguments for " + name;
            }
            return params;
        }
    }
}

function errorUsage(v, msg) {
    console.log("Error: Missing arguments for " + v);
    for (var i = 0; i < COMMAND_LIST.length; i++) { 
        var c = COMMAND_LIST[i][0];
        if (v == c) {
            console.log("Usage: " + COMMAND_LIST[i][2]);
        }
    }
}

function queryUser() {
    console.log("Choose user ...");
    prompt.setPrompt("user: ", 6);
    prompt.prompt();
}

/////////////// commands

function doHelp(v) {
    if (!v) {
        console.log('Available commands');
        for (var i = 0; i < COMMAND_LIST.length; i++) { 
            var c = COMMAND_LIST[i][0];
            console.log(c + "                    ".substring(0, 20 - c.length) + COMMAND_LIST[i][1]);
        }
    } else {
        var found = false;
        for (var i = 0; i < COMMAND_LIST.length; i++) { 
            var c = COMMAND_LIST[i][0];
            if (v == c) {
                console.log(COMMAND_LIST[i][1]);
                console.log("Usage: " + COMMAND_LIST[i][2]);
                found = true;
            }
        }
        if (!found) {
            throw "Uknown command: " + v;
        }
    }
}

function doGreet(v) {
    if (!v) {
        errorUsage("greet");
        return;
    }
    var req = atmosphere.util.stringifyJSON({ "id": getNextId(), "method": "POST", "path": "/v1/greet/" + user, "type": "application/json"})+atmosphere.util.stringifyJSON({ "name": v[0], "text": v[1]});

    if (trace) {
        console.log("TRACE: sending ", req);
    }
    subSocket.push(req);
}

function doStatus(v) {
    if (!v) {
        errorUsage("status");
        return;
    }
    var req = atmosphere.util.stringifyJSON({ "id": getNextId(), "method": "GET", "path": "/v1/greet/" + v[0]});

    if (trace) {
        console.log("TRACE: sending ", req);
    }
    subSocket.push(req);
}

function doSummary(v) {
    var req = atmosphere.util.stringifyJSON({ "id": getNextId(), "method": "GET", "path": "/v1/greet"});

    if (trace) {
        console.log("TRACE: sending ", req);
    }
    subSocket.push(req);
}

function doSubscribe(v) {
    var req = atmosphere.util.stringifyJSON({ "id": getNextId(), "method": "GET", "path": "/v1/subscribe/" + user});

    if (trace) {
        console.log("TRACE: sending ", req);
    }
    subSocket.push(req);
}

function doUnsubscribe(v) {
    var req = atmosphere.util.stringifyJSON({ "id": getNextId(), "method": "DELETE", "path": "/v1/unsubscribe/" + v[0]});

    if (trace) {
        console.log("TRACE: sending ", req);
    }
    subSocket.push(req);
}

function doPing(v) {
    var req = atmosphere.util.stringifyJSON({ "id": getNextId(), "method": "GET", "path": "/v1/ping"});

    if (trace) {
        console.log("TRACE: sending ", req);
    }
    subSocket.push(req);
}

function doEcho(v) {
    if (!v) {
        errorUsage("echo");
        return;
    }
    var req = atmosphere.util.stringifyJSON({ "id": getNextId(), "method": "POST", "path": "/v1/echo", "type": "text/plain"})+v[0];

    if (trace) {
        console.log("TRACE: sending ", req);
    }
    subSocket.push(req);
}


function doQuit() {
    subSocket.close();
    process.exit(0);
}


request.onOpen = function(response) {
    // workaround of onOpen being invoked twice when some data is pushed from this method
    if (isopen) return;
    isopen = true;
    console.log('Connected using ' + response.transport);
    subSocket.push(JSON.stringify({ "version": PROTOCOL_VERSION}));
    prompt.setPrompt(userprompt, 2);
    prompt.prompt();
};

request.onReopen = function(response) {
    // workaround of onOpen being invoked twice when some data is pushed from this method
    if (isopen) return;
    isopen = true;
    console.log('Reopened using ' + response.transport);
    subSocket.push(JSON.stringify({ "version": PROTOCOL_VERSION}));
    prompt.setPrompt(userprompt, 2);
    prompt.prompt();
}

request.onReconnect = function(response) {
    console.log("Reconnecting ...");
}

request.onMessage = function (response) {
    var message = response.responseBody;
    var jpart;
    var data;
    var json;
    //FIXME use a better logic to determine the mode
    var messageparts = splitMessage(message);
    jpart = messageparts[0];
    data = messageparts[1];
    try {
        json = JSON.parse(jpart);
    } catch (e) {
        console.log('Invalid response: ', message);
        return;
    }
    if (json.version) {
        if (trace) {
            console.log("TRACE: received " + message);
            prompt.setPrompt(userprompt, 2);
            prompt.prompt();
        }
        if (json.error) {
            console.log("Error:" + json.error);
            doQuit();
        }
    } else {
        if (trace) {
            console.log("TRACE: received " + message);
            prompt.setPrompt(userprompt, 2);
            prompt.prompt();
        }

        if (json.error_code) {
            console.log(jpart);
        } else if (json.id) {
            console.log("res" + json.id + ":", message);
        } else {
            // no id supplied in the response, so just write the plain result
            console.log("res*" + ":", message)
        }
    }
    prompt.setPrompt(userprompt, 2);
    prompt.prompt();
};

request.onClose = function(response) {
    console.log("Closed");
    isopen = false;
}

request.onError = function(response) {
    console.log("Sorry, something went wrong: " + response.responseBody);
};

var transport = null;
var subSocket = null;
var user = null;
var userprompt = "> ";

queryUser();

prompt.
on('line', function(line) {
    try {
        var msg = line.trim();
        if (user == null) {
            user = msg
            userprompt = user + userprompt;
            subSocket = atmosphere.subscribe(request);
            console.log("Connecting using " + request.transport + " ...");
            setTimeout(function() {
                if (!isopen) {
                    console.log("Unable to open a connection. Terminated.");
                    process.exit(0);
                }
            }, 3000);
        } else if (msg.length == 0) {
            doHelp();
        } else if (msg.indexOf("ping") == 0) {
            doPing();
        } else if (msg.indexOf("echo") == 0) {
            doEcho(getArgs("echo", msg, 1));
        } else if (msg.indexOf("greet") == 0) {
            doGreet(getArgs("greet", msg, 2));
        } else if (msg.indexOf("help") == 0) {
            doHelp(getArgs("help", msg, 0));
        } else if (msg.indexOf("status") == 0) {
            doStatus(getArgs("status", msg, 1));
        } else if (msg.indexOf("summary") == 0) {
            doSummary();
        } else if (msg.indexOf("subscribe") == 0) {
            doSubscribe();
        } else if (msg.indexOf("unsubscribe") == 0) {
            doUnsubscribe(getArgs("unsubscribe", msg, 1));
        } else if (msg.indexOf("quit") == 0) {
            doQuit();
        } else {
            throw "Uknown command: " + msg;
        }
    } catch (err) {
        console.log("Error: " + err);
    }
    prompt.setPrompt(userprompt, 2);
    prompt.prompt();
}).

on('close', function() {
    doQuit();
});
