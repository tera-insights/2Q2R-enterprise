/// <reference path="../../typings/index.d.ts" />

import { IAuthenticationItem, IAuthenticationSocket, Authentications } from '../services/Activity';

export class DashboardCtrl {

    private authentications: IAuthenticationItem[];
    private filteredAuthentications: IAuthenticationItem[];
    private selectedAuthentications: IAuthenticationItem[];

    private appCount: number = Math.floor(Math.random() * 100);
    private serverCount: number = Math.floor(Math.random() * 2000);
    private adminCount: number = Math.floor(Math.random() * 300);
    private userCountInThousands: number = Math.floor(Math.random() * 100);

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

    private filterOpen: boolean = false;
    private filters: string[] = [
        "time",
        "name",
        "country"
    ];
    private selectedFilter: string;
    private filterValue: string;

    private orderBy: string;
    private pagination = {
        pageLimit: 11,
        page: 1
    }

    applyFilters() {
        if (this.selectedFilter && this.filterValue) {
            this.filteredAdmins = this.admins.filter((admin: IAdminItem): boolean => {
                return admin[this.selectedFilter].includes(this.filterValue);
            });
        } else {
            this.filteredAdmins = this.admins;
        }
    }

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