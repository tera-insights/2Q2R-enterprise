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

    private appCount: number = Math.floor(Math.random() * 100);
    private serverCount: number = Math.floor(Math.random() * 2000);
    private adminCount: number = Math.floor(Math.random() * 300);
    private userCountInThousands: number = Math.floor(Math.random() * 100);

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

            var ctx = document.getElementById("chart");
            var myChart = new Chart(ctx, {
                type: 'bar',
                data: {
                    labels: ["Red", "Blue", "Yellow", "Green", "Purple", "Orange"],
                    datasets: [{
                        label: '# of Votes',
                        data: [12, 19, 3, 5, 2, 3],
                        backgroundColor: [
                            'rgba(255, 99, 132, 0.2)',
                            'rgba(54, 162, 235, 0.2)',
                            'rgba(255, 206, 86, 0.2)',
                            'rgba(75, 192, 192, 0.2)',
                            'rgba(153, 102, 255, 0.2)',
                            'rgba(255, 159, 64, 0.2)'
                        ],
                        borderColor: [
                            'rgba(255,99,132,1)',
                            'rgba(54, 162, 235, 1)',
                            'rgba(255, 206, 86, 1)',
                            'rgba(75, 192, 192, 1)',
                            'rgba(153, 102, 255, 1)',
                            'rgba(255, 159, 64, 1)'
                        ],
                        borderWidth: 1
                    }]
                },
                options: {
                    scales: {
                        yAxes: [{
                            ticks: {
                                beginAtZero: true
                            }
                        }]
                    }
                }
            });
        }, 100);
    }

}