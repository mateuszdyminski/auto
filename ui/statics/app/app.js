'use strict';

var FC = angular.module('FC', ['ngRoute', 'leaflet-directive', 'ui.bootstrap', 'toastr'])
    .config(function($routeProvider) {
        $routeProvider
            .when('/crashes', {
                templateUrl: 'app/views/crashes.html',
                controller: 'FlightsCtrl'
            })
            .otherwise({
                redirectTo: '/crashes'
            });
    });