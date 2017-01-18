import { AdminSrvc } from '../services/AdminSrvc';
import { AddAdminCtrl } from './modals/AddAdminCtrl';

export class AdminsCtrl {

    private admins: any[];
    private filteredAdmins: any[];
    private selectedAdmins: any[];

    private filterOpen: boolean = false;
    private filters: string[] = [
        "name",
        "status",
        "email",
        "role"
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
            this.filteredAdmins = this.admins.filter((admin: any): boolean => {
                return admin[this.selectedFilter].includes(this.filterValue);
            });
        } else {
            this.filteredAdmins = this.admins;
        }
    }

    refresh() {
        this.admins = this.AdminSrvc.get('');
    }

    deleteSelected() {
    }

    referAdmin() {
        this.$mdDialog.show({
            controller: AddAdminCtrl,
            controllerAs: 'cMod',
            templateUrl: 'views/modals/AddAdmin.html',
            clickOutsideToClose: true
        }).then((error?: Error) => {
            // TODO: refer admin
            if(!error) {
                this.$mdToast.showSimple('Admin successfully registered!');
            } else {
                this.$mdToast.showSimple(error.message);
            }
        }, () => {
            // user canceled, do nothing
        });
    }

    static $inject = [
        'AdminSrvc',
        '$mdDialog',
        '$mdToast'
    ];

    constructor(
        private AdminSrvc: AdminSrvc,
        private $mdDialog: ng.material.IDialogService,
        private $mdToast: ng.material.IToastService
    ) {
        this.refresh();
        this.applyFilters()
    }

}