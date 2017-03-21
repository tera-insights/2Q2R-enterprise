"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var DeleteServersCtrl = (function () {
    function DeleteServersCtrl($mdDialog) {
        this.$mdDialog = $mdDialog;
    }
    DeleteServersCtrl.prototype.accept = function () {
        this.$mdDialog.hide();
    };
    DeleteServersCtrl.prototype.cancel = function () {
        this.$mdDialog.cancel();
    };
    return DeleteServersCtrl;
}());
DeleteServersCtrl.$inject = [
    '$mdDialog'
];
exports.DeleteServersCtrl = DeleteServersCtrl;

//# sourceMappingURL=DeleteServersCtrl.js.map
