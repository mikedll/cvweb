
document.addEventListener("DOMContentLoaded", () => {

  console.log(`main.js executing`);
  
  const pastResults = document.querySelector('.past-results table');
  if(pastResults !== null) {
    if(window.results === '') {
      pastResults.appendChild(document.createTextNote("No past results."));
      return;
    }
    
    const results = JSON.parse(window.results);

    if(results.length === 0) {
      const p = document.createElement('p');
      p.appendChild(document.createTextNode('0 results found.'));
      pastResults.closest('div').appendChild(p);
      return;
    }
    
    // console.log(`found ${results.length} results`);
    const tbody = pastResults.querySelector('tbody');
    results.forEach((result) => {
      console.log("result: ", result.uuid, result.createdAt);
      const tr = document.createElement('tr');
      const uuidCell = document.createElement('td');
      const a = document.createElement('a');
      a.appendChild(document.createTextNode(result.uuid));
      a.href = `/requests/${result.uuid}`;
      uuidCell.appendChild(a)

      const createdAtCell = document.createElement('td');
      createdAtCell.appendChild(document.createTextNode(result.createdAt));

      tr.appendChild(uuidCell);
      tr.appendChild(createdAtCell);
      
      tbody.appendChild(tr);
    });    
  } else {
    // console.log(`no past results node found`);    
  }

});
