import { Admins, IAdminItem, IAdminResource } from '../services/Admins';

export class AdminsCtrl {

    private admins: IAdminItem[];
    private filteredAdmins: IAdminItem[];
    private selectedAdmins: IAdminItem[];

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
            this.filteredAdmins = this.admins.filter((admin: IAdminItem): boolean => {
                return admin[this.selectedFilter].includes(this.filterValue);
            });
        } else {
            this.filteredAdmins = this.admins;
        }
    }

    refresh() {
        this.admins = this.Admins.resource.query();
    }

    static $inject = [
        'Admins'
    ];

    constructor(
        private Admins: Admins
    ) {
        this.refresh();
        this.applyFilters()
    }

}