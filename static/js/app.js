$(function () {
    $.get("/data", function (data) {
        console.log(data);

        $("#username-text .data").html(data.username);
        $("#follower-count .data").html(data.follower_count);

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
                legend: {
                    position: 'top',
                },
                title: {
                    display: true,
                    text: 'Daily Follower Count'
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
                legend: {
                    position: 'top',
                },
                title: {
                    display: true,
                    text: 'Daily Follower Events'
                }
            }
        });

        loadUsers($("#follower-list"), data.recent_follower_list);
        loadUsers($("#unfollower-list"), data.recent_unfollower_list);

    });
});


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
                " <b>Post:</b> " + u.post_count + ")"),
            $("<div class='user-info-desc'>").text(u.description),
        );

        var $tr = $("<tr class='user-row'>").append(
            $("<td class='user-img'>").html("<img src='" + u.profile_image + "'/>"),
            $("<td class='user-info'>").append($info)
        ).appendTo(tbl);
    });


}