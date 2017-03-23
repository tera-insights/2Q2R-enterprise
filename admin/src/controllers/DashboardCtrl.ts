import { StatsSrvc } from '../services/StatsSrvc';

export class DashboardCtrl {

    // TODO: add authentication socket to listen for new logins
    // private authentications: IAuthenticationItem[];
    // private filteredAuthentications: IAuthenticationItem[];
    // private selectedAuthentications: IAuthenticationItem[];

    private appCount: number = Math.floor(Math.random() * 100);
    private serverCount: number = Math.floor(Math.random() * 2000);
    private adminCount: number = Math.floor(Math.random() * 300);
    private userCountInThousands: number = Math.floor(Math.random() * 100);

    private communicationID: number;

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
        
    }

    onServerNotification(msg: any[]) {
        // update dashboard
    }

    static $inject = [
        'StatsSrvc',
        '$timeout'
    ];

    constructor(
        private StatsSrvc: StatsSrvc,
        private $timeout: ng.ITimeoutService
    ) {
        this.communicationID = this.StatsSrvc.subscribe(['registration', 'authentication'], this.onServerNotification);
        this.$timeout(() => {
            var map = L.map('map').setView([this.lConfig.center.lat, this.lConfig.center.lng], this.lConfig.center.zoom);
            L.tileLayer(this.lConfig.tileLayer, {
                maxZoom: 19,
                id: 'open.street.map',
                attribution: this.lConfig.tileAttrib
            }).addTo(map);
        }, 100);
    }

    destructor() {
        this.StatsSrvc.unsubscribe([this.communicationID]);
    }

}