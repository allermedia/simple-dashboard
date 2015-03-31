var Dashboard = React.createClass({
  getInitialState: function() {
    return {data: []};
  },
  loadDashboard: function() {
    $.ajax({
      url: this.props.url,
      datatype: 'json',
      success: function(data) {
        this.setState({data: data});
      }.bind(this),
      error: function(xhr, status, err) {
        console.error(this.props.url, status, err.toString());
      }.bind(this)
    });
  },
  componentDidMount: function() {
    this.loadDashboard()
    ws = new WebSocket("ws://"+window.location.host+"/receive");
    console.log("WebSocket opened")
    ws.onmessage = function(evt) {
      var entry = JSON.parse(evt.data)
      var entries = this.state.data
      entries.unshift(entry)

      for (var i = 1; i < entries.length; i++) {
        if (entries[i].name == entry.name) {
          entries.splice(i, 1)
          break;
        }
      }

      this.setState({data: entries})
    }.bind(this)
  },
  render: function() {
    return (
      <div>
        <h1>Build dashboard</h1>
        <div id="entry-list">
        <BuildEntryList data={this.state.data} />
        </div>
      </div>
    )
  }
});

var BuildEntryList = React.createClass({
  render: function() {
    var entries = this.props.data.map(function (entry) {
      return (
        <Entry status={entry.status} name={entry.name} />
      );
    });
    return (
      <div className="entries">
      {entries}
      </div>
    );
  }
});

var Entry = React.createClass({ 
  render: function() {
    return (
      <div className={this.props.status}>
      <h2>{this.props.name}</h2>
      </div>
    );
  }
});



React.render(
  <Dashboard url="/entry/build" />,
  document.getElementById('build-dashboard')
);