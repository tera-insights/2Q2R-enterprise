var keyType = "2q2r"; //the keytype selected
var hasU2FExt = false;

addState("keytype", "registerKeyList", function () {
    return {
        keys: data.keyTypes.map(makeKeyInfo)
    }
}, function (sel) {
    data.keyTypes.forEach(function (key) {
        $("#" + key + "-cb").click(function (event) {
            $("div", "#keyInfo").hide();
            $("#" + key + "-ki").show();
        });
    });

    $("#continue", sel).click(
        function (event) {
            keyType = $('input[name=keyradio]:checked', '#form-kt').val();
            selectState(keyType + "-generate");
        }
    )

    $("#trigger-passwd").click(
        function (event) { $('#modal-passwd').openModal(); });

    $("#trigger-priv").click(
        function (event) { $('#modal-priv').openModal(); });


});

addState("2q2r-generate", "2q2rRegister",
    function () {
        return {}
    },
    function (sel) {
        $("#trigger-challenge").click(
            function (event) { $('#modal-challenge').openModal(); });

        $("#trigger-notify").click(
            function (event) { $('#modal-notify').openModal(); })

        $("#continue", sel).click(
            function (event) {

            }
        );
        $("#back", sel).click(
            function (event) {
                selectState('keytype');
            }
        );
        $("#qrcode", sel).empty();
        $("#qrcode", sel).qrcode({
            text: "R " + data.challenge + " " + data.infoUrl + " " + data.userID,
            width: 192, height: 192
        });

        var disp = displayableChallenge(data.challenge);
        $("#challenge-big", sel).text(disp.big);
        $("#challenge-small", sel).text(disp.small);

        $.get(data.waitUrl, function (res) {
            if (res.successful) {
                $("#qrcode", sel).empty();
                $("#qrcode", sel).append("<img src=\"/img/check.png\" alt=\"Registration\n" +
                    "Successful!\" id=\"successImage\" style=\"width: 174px; height: 174px;\">");
            }
        }).fail(function (jqXHR, textStatus) {
            $("#qrcode", sel).empty();
            $("#qrcode", sel).append("<img src=\"/img/timeout.png\" alt=\"Registration\n" +
                "timed out.\" id=\"timeoutImage\" style=\"width: 174px; height: 174px;\">");
        });

    });


addState('u2f-generate', 'u2fRegister',
    function () {
        return {
            windowUrl: window.location + "?k=u2f",
            ext: hasU2FExt
        }
    }, function (sel) {
        var registerRequests = [{ version: "U2F_V2", challenge: data.challenge }];
        console.log("Sending U2F registratoin");
        u2f.register(data.baseUrl, registerRequests, [], function (reply) {
            console.log("Registration data: ", reply);

            if (reply.errorCode) {
                $("#u2f-window", sel).show();
                $("#u2f-msg").hide();
                return;
            }

            // add more info to reply to get it ready to give to 2Q2R server
            reply.appID = data.appId;
            reply.challenge = data.challenge;
            reply.deviceName = "YubiKey";
            reply.type = "u2f";
            $.postJSON(data.registerUrl,
                { successful: true, data: reply },
                function (res) {
                    if (res.successful)
                        console.log("Succesful: ", res);
                    else
                        console.log("Error: ", res);
                }).fail(function (jqXHR, textStatus) {
                    console.log("Error: ", jqXHR.status);
                });
        });
    }
);

$(document).ready(function () {
    init({
        title: 'Device Registration'
    });

    var query = window.location.search.substr(1);
    if (query === "k=u2f") {
        selectState("u2f-generate");
        $('html').addClass("window");
    }
    else {
        // start with key type selection
        selectState("keytype");
    }

    hasU2FExt =
        bowser.check({ firefox: "41" }) && u2f.sign && !u2f.getApiVersion ||
        bowser.check({ chrome: "41" })
        ;

    console.log("Has ext: ", hasU2FExt);

});