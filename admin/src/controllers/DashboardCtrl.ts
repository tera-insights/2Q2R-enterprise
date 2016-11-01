/// <reference path="../../typings/index.d.ts" />

export class DashboardCtrl {

    private lConfig = {
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

    static $inject = [
        '$timeout'
    ];

    constructor(
        private $timeout: ng.ITimeoutService
    ) {
        this.$timeout(() => {
            var map = L.map('map').setView([this.lConfig.center.lat, this.lConfig.center.lng], this.lConfig.center.zoom);
            L.tileLayer(this.lConfig.tileLayer, {
                maxZoom: 19,
                id: 'open.street.map',
                attribution: this.lConfig.tileAttrib
            }).addTo(map);
        }, 100);
    }

}