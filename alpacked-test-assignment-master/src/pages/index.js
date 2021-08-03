import { graphql } from "gatsby"
import React from "react"
import Cards from "../components/Cards"
import Hero from "../components/Hero"
import Layout from "../layouts/Layout"
import Newsletter from "../components/Newsletter"
import SiteMetadata from "../components/SiteMetadata"

const IndexPage = ({ data }) => {
  return (
    <div>
      {console.log(data)}
      {data.posts.nodes.map(post =>(
        <div key={post.id}>
          <p>{post.title}</p>
          <img src={post.image.file.url}/>
        </div>
      ))}
    </div>
  )
}

export default IndexPage

export const query = graphql`
  query PostsQuery{
    posts: allContentfulPost{
      nodes{
        id
        title
        image{
          file{
            url
          }
        }
      }
    }
  }
`
