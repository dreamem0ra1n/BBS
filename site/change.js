var fs = require('fs')
var data = fs.readFileSync('./oldForum.json')
data = JSON.parse(data)
var res = data.filter((item)=>item.type==='group').map((item) => {
    let newR = {};
    newR.type = "group"
    newR.fid = item.fid
    newR.name = item.name
    newR.fdn = []
    return newR
})
data.forEach((item)=>{
    res.forEach((group)=>{
        if(group.fid===item.fup)
        {
            group.fdn.push({
                fname:item.name,
                fid:item.fid,
                type:"forum",
                fdn:[]
            })
        }
    })
})
res.forEach((group)=>{
    group.fdn.forEach((forum)=>{
        data.forEach((item)=>{
            if(forum.fid===item.fup)
            {
                forum.fdn.push({
                    fname:item.name,
                    fid:item.fid,
                    type:"sub"
                })
            }
        })
    })
})
res-res.sort((a,b)=>{
    return a.fid-b.fid
})
fs.writeFileSync('./new.json', JSON.stringify(res))
