$(function () {
    if ($("#numbers-section").length) {
        runQuery();
    }
});

function runQuery() {
    $.get("/data", function (data) {
        console.log(data);

        // numbers
        $("#follower-count .data").text(data.user.followers_count);
        $("#following-count .data").text(data.user.following_count);
        $("#follower-gained-count .data").text(data.recent_follower_count);
        $("#follower-lost-count .data").text(data.recent_unfollower_count);
        $("#favorites-count .data").text(data.user.fave_count);

        $("#meta-panel").html("Account: <b>" + data.user.username + "</b>" +
            " | Time period: <b>Last " + data.meta.num_days_period + "days</b>" +
            " | Updated on: <b>" + toLongTime(data.user.updated_on) + "</b>" +
            " | <a href='/auth/logout'>Log out</a>"
        );


        // follower count chart
        var followerChart = new Chart($("#follower-count-series")[0].getContext("2d"), {
            type: 'bar',
            data: {
                labels: Object.keys(data.follower_count_series),
                datasets: [{
                    label: 'Count',
                    data: Object.values(data.follower_count_series),
                    backgroundColor: 'rgba(54, 162, 235, 0.2)',
                    borderColor: 'rgba(54, 162, 235, 1)',
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                title: {
                    display: true,
                    text: 'Daily Follower Count'
                },
                legend: {
                    display: false
                },
                scales: {
                    yAxes: [
                        {
                            ticks: {
                                stepSize: 5
                            }
                        }
                    ]
                }
            }
        });




        // follower count chart
        var followerChart = new Chart($("#follower-event-series")[0].getContext("2d"), {
            type: 'bar',
            data: {
                labels: Object.keys(data.followed_event_series),
                datasets: [{
                    label: 'Followed',
                    data: Object.values(data.followed_event_series),
                    backgroundColor: 'rgba(0,255,0,0.3)',
                    borderColor: 'rgba(0,255,0,1)',
                    borderWidth: 1
                }, {
                    label: 'Unfollowed',
                    data: Object.values(data.unfollowed_event_series),
                    backgroundColor: 'rgba(255,0,0,0.3)',
                    borderColor: 'rgba(255,0,0,1)',
                    borderWidth: 1
                }]
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                title: {
                    display: true,
                    text: 'Daily Follower Events'
                },
                legend: {
                    display: false
                },
                scales: {
                    yAxes: [
                        {
                            ticks: {
                                beginAtZero: true
                            }
                        }
                    ]
                }
            }
        });

        loadUsers($("#follower-list"), data.recent_follower_list);
        loadUsers($("#unfollower-list"), data.recent_unfollower_list);

    });
}

function toShortDate(v) {
    var ts = new Date(v)
    return ts.toISOString().substring(0, 10)
}

function toLongTime(v) {
    var ts = new Date(v)
    return ts.toTimeString()
}



function loadUsers(tbl, list) {

    $.each(list, function (i, u) {
        // console.log("User[" + i + "] ID: " + u.id);

        /*
        created_at: "2018-06-02T00:00:00Z"
        description: "Future Global #Halal Marketplace"
        fave_count: 0
        followers_count: 251
        following_count: 0
        id: "1003044630592712704"
        lang: ""
        location: "London, England"
        name: "YouHalal.com"
        post_count: 102858
        profile_image: "https://pbs.twimg.com/profile_images/1017890735100780544/TLnzh-nB_normal.jpg"
        time_zone: ""
        username: "YouHalal"
        */

        var $info = $("<div class='user-info-detail'>").append(
            $("<div class='user-info-name'>").html("<b>" + u.username + "</b> - " + u.name +
                " (<b>Loc:</b> " + u.location +
                " <b>Follow:</b> " + u.followers_count + "/" + u.following_count +
                " <b>Post:</b> " + u.post_count +
                " <b>On:</b> " + toShortDate(u.event_at) + ")"),
            $("<div class='user-info-desc'>").text(u.description),
        );

        var $tr = $("<tr class='user-row'>").append(
            $("<td class='user-img'>").html("<img src='" + u.profile_image + "'/>"),
            $("<td class='user-info'>").append($info)
        ).appendTo(tbl);
    });


}