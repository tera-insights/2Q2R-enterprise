var keyIndex = 0; // the key selected

addState("keyselect", "deleteKeyList",
    function () {
        return {
            keys: data.keys
        };
    },
    function (sel) {
        $("#continue", sel).click(
            function (event) {
                keyIndex = $('input[name=keyradio]:checked', '#form-kt').val();

                console.log("Key: ", keyIndex, data.keys[keyIndex]);

                $.delete(data.deleteUrl + '/' + data.keys[keyIndex].keyID)
                    .always(function (jqXHR, textStatus) {
                        if (jqXHR.status === 200) {
                            selectState("deleted");
                        }

                        else {
                            selectState("error");
                        }

                    });

            }
        )
    });

addState("deleted", "deleteMessage",
    function () {
        return {
            key: data.keys[keyIndex]
        }
    },
    function (event) {
        keyIndex = $('input[name=keyradio]:checked', '#form-kt').val();
        var keyType = data.keys[keyIndex].type;
        selectState(keyType + "-login");
    })
$(document).ready(function () {
    init({
        title: 'Delete Device'
    });

    // start with key type selection
    selectState("keyselect");

});
