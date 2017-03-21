"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var L = require("leaflet");
var DashboardCtrl = (function () {
    function DashboardCtrl(StatsSrvc, $timeout) {
        var _this = this;
        this.StatsSrvc = StatsSrvc;
        this.$timeout = $timeout;
        this.appCount = Math.floor(Math.random() * 100);
        this.serverCount = Math.floor(Math.random() * 2000);
        this.adminCount = Math.floor(Math.random() * 300);
        this.userCountInThousands = Math.floor(Math.random() * 100);
        this.lConfig = {
            tileLayer: 'http://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png',
            tileAttrib: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>',
            maxZoom: 14,
            path: {
                weight: 10,
                color: '#800000',
                opacity: 1
            },
            center: {
                lat: 51.505,
                lng: -0.09,
                zoom: 8
            }
        };
        this.filterOpen = false;
        this.filters = [
            "time",
            "name",
            "country"
        ];
        this.pagination = {
            pageLimit: 11,
            page: 1
        };
        this.communicationID = this.StatsSrvc.subscribe(['registration', 'authentication'], this.onServerNotification);
        this.$timeout(function () {
            var map = L.map('map').setView([_this.lConfig.center.lat, _this.lConfig.center.lng], _this.lConfig.center.zoom);
            L.tileLayer(_this.lConfig.tileLayer, {
                maxZoom: 19,
                id: 'open.street.map',
                attribution: _this.lConfig.tileAttrib
            }).addTo(map);
        }, 100);
    }
    DashboardCtrl.prototype.applyFilters = function () {
    };
    DashboardCtrl.prototype.onServerNotification = function (msg) {
    };
    DashboardCtrl.prototype.destructor = function () {
        this.StatsSrvc.unsubscribe([this.communicationID]);
    };
    return DashboardCtrl;
}());
DashboardCtrl.$inject = [
    'StatsSrvc',
    '$timeout'
];
exports.DashboardCtrl = DashboardCtrl;

//# sourceMappingURL=DashboardCtrl.js.map
