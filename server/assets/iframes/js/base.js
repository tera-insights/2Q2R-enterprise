var state = "";
var states = {};

/**
 * Auxiliary function to create key info from key type
 */
function makeKeyInfo(type) {
    switch (type) {
        case "2q2r":
            return {
                value: "2q2r", name: "2Q2R App", icon: "img/2q2r.png"
            };

        case "u2f":
            return {
                value: "u2f", name: "FIDO U2F Device", icon: "img/u2f.png"
            };

        case "software":
            return {
                value: "software", name: "Software Key"
            };

        case "bypass":
            return {
                value: "bypass", name: "Bypass Key"
            }

        default:
            return {}
    }
}

/**
 * This function must be called once the page is loaded
 */
function init(data) {
    // JQuery Shim to post JSON
    jQuery["postJSON"] = function (url, data, callback) {
        // shift arguments if data argument was omitted
        if (jQuery.isFunction(data)) {
            callback = data;
            data = undefined;
        }

        return jQuery.ajax({
            url: url,
            type: "POST",
            contentType: "application/json; charset=utf-8",
            dataType: "json",
            data: JSON.stringify(data),
            success: callback
        });
    };

    // Render the main template and attach it to the body
    dust.render('main', data, function (err, out) {
        $('body').html(out);
        // TODO: hook up global actions
    })
}

/**
 * Function to add a state
 * @name {string} State name
 * @template {string} Template name
 * @init {Function} State init code. Triggered when state changes. Must produce data for template
 * @actions {Function} State action hookup function. Must attach actions after template rendered.
 */
function addState(name, template, init, actions) {
    states[name] = { template: template, init: init, actions: actions };
}


/**
 * Auxiliary function to initialize the semantic objects. This ensures
 * that semantic things come "alive".
 */
function semanticInit() {
    // rename
}

/**
 * Function to switch states. This controls the <section> active
 * Call the state init with state selector as argument as well
 */
function selectState(state) {
    var stateSel = $('#state');

    var stateInfo = states[state];
    if (!stateInfo)
        return;
    var data = (stateInfo.init) ? stateInfo.init() : {};
    if (stateInfo.template)
        dust.render(stateInfo.template, data, function (err, out) {
            if (err)
                stateSel.html("<h1>Error: " + JSON.stringify(err) + " </h1>");
            else {
                stateSel.html(out);
                if (stateInfo.actions)
                    stateInfo.actions(stateSel);
                semanticInit();
            }
        });
    else
        stateSel.html("<h1>State " + state + " not implemented </h1>");
}

/**
 * Auxiliary function to extract big and small displayable text out of challenges
 * 
 * @ch {string} the challenge to process
 * @returns { big: string, small: string }
 */
function displayableChallenge(ch) {
    // compute small text
    var small = ch.substr(8, 4);
    for (var i = 12; (i < ch.length && i < 28); i += 4) {
        small += " " + ch.substr(i, 4);
    }
    return {
        big: ch.substr(0, 4) + " " + ch.substr(4, 4),
        small: small
    }
}