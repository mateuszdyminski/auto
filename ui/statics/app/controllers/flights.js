'use strict';

angular.module('FC').controller('FlightsCtrl', function($scope, $timeout) {
    $scope.flights = [];
    $scope.markers = [];

    var ws = new WebSocket('ws://192.168.99.100:32090/wsapi/ws');

    ws.onmessage = function(event) {
        $timeout(function() {
            $scope.$apply(function(){
                var crash = JSON.parse(event.data);
                $scope.flights.push(crash);
                var marker = {
                    lng: crash.locationGPS.lon,
                    lat: crash.locationGPS.lat,
                    message: JSON.stringify(crash, null, 2),
                    focus: true
                };
                $scope.markers.push(marker);
    
                if ($scope.flights.length > 300) {
                    $scope.flights.pop();
                    $scope.markers.pop();
                }    
            });
        });
    };
});